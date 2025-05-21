package main

import (
    "bufio"
    "database/sql"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/fsnotify/fsnotify"
)

func connectDB() (*sql.DB, error) {
    dsn := "tacacsdb:tacacsdb#123db@tcp(tacacsdb:3306)/tacacsdb"
    var db *sql.DB
    var err error
    for {
        db, err = sql.Open("mysql", dsn)
        if err != nil {
            fmt.Println("Erro ao conectar ao banco:", err)
        } else {
            err = db.Ping()
            if err == nil {
                fmt.Println("Banco de dados conectado com sucesso!")
                break
            } else {
                fmt.Println("Banco de dados não disponível. Tentando novamente...")
            }
        }
        time.Sleep(1 * time.Second)
    }
    return db, nil
}

func insertLogData(db *sql.DB, timestamp, ip, username, interfaceName, clientIP, action string) error {
    query := `INSERT INTO logs (timestamp, ip, username, interface, client_ip, action) VALUES (?, ?, ?, ?, ?, ?)`
    _, err := db.Exec(query, timestamp, ip, username, interfaceName, clientIP, action)
    return err
}

func parseLogLine(line string) (string, string, string, string, string, string, error) {
    parts := strings.Split(line, "\t")
    if len(parts) < 6 {
        return "", "", "", "", "", "", fmt.Errorf("linha de log inválida")
    }

    rawTimestamp := parts[0]
    // Parse do timestamp com fuso horário
    parsedTime, err := time.Parse("2006-01-02 15:04:05 -0700", rawTimestamp)
    if err != nil {
        return "", "", "", "", "", "", fmt.Errorf("erro ao parsear timestamp: %v", err)
    }
    formattedTimestamp := parsedTime.Format("2006-01-02 15:04:05")

    ip := parts[1]
    username := parts[2]
    interfaceName := parts[3]
    clientIP := parts[4]
    action := strings.Join(parts[5:], " ")

    return formattedTimestamp, ip, username, interfaceName, clientIP, action, nil
}

func monitorFile(filePath string, db *sql.DB) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    file, err := os.Open(filePath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Move para o final do arquivo inicialmente
    offset, _ := file.Seek(0, os.SEEK_END)

    fmt.Println("Monitorando o arquivo:", filePath)
    err = watcher.Add(filePath)
    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                newOffset, err := processNewLines(filePath, offset, db)
                if err != nil {
                    fmt.Println("Erro ao processar novas linhas:", err)
                } else {
                    offset = newOffset
                }
            }
        case err := <-watcher.Errors:
            fmt.Println("Erro de monitoramento:", err)
        }
    }
}

func processNewLines(filePath string, offset int64, db *sql.DB) (int64, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return offset, err
    }
    defer file.Close()

    _, err = file.Seek(offset, 0)
    if err != nil {
        return offset, err
    }

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }

        fmt.Printf("Linha lida do log: %q\n", line)

        timestamp, ip, username, interfaceName, clientIP, action, err := parseLogLine(line)
        if err != nil {
            fmt.Println("Erro ao processar linha do log:", err)
            continue
        }

        err = insertLogData(db, timestamp, ip, username, interfaceName, clientIP, action)
        if err != nil {
            fmt.Println("Erro ao inserir dados no banco:", err)
        }
    }

    // Retorna o novo offset atual após leitura
    newOffset, err := file.Seek(0, os.SEEK_CUR)
    if err != nil {
        return offset, err
    }

    return newOffset, nil
}

func main() {
    db, err := connectDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    filePath := "/applog/tacacs.log"
    monitorFile(filePath, db)
}
