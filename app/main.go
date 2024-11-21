package main

import (
    "bufio"
    "database/sql"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    _ "github.com/go-sql-driver/mysql" // Importa o driver MySQL
    "github.com/fsnotify/fsnotify"
)

// Conectar ao banco de dados MySQL, com checagem de disponibilidade
func connectDB() (*sql.DB, error) {
    // Substitua pelos detalhes do seu banco de dados
    dsn := "tacacsdb:tacacsdb#123db@tcp(tacacsdb:3306)/tacacsdb"

    // Tenta abrir a conexão e fazer o ping
    var db *sql.DB
    var err error

    // Loop até que a conexão seja bem-sucedida
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

        // Aguarda 1 segundo antes de tentar novamente
        time.Sleep(1 * time.Second)
    }

    return db, nil
}

// Inserir dados no banco de dados MySQL
func insertLogData(db *sql.DB, timestamp, ip, username, interfaceName, clientIP, action string) error {
    // Query de inserção no banco de dados
    query := `INSERT INTO logs (timestamp, ip, username, interface, client_ip, action) 
              VALUES (?, ?, ?, ?, ?, ?)`
    _, err := db.Exec(query, timestamp, ip, username, interfaceName, clientIP, action)
    return err
}

// Parse da linha do log
func parseLogLine(line string) (string, string, string, string, string, string, error) {
    // Quebrar a linha em partes usando o delimitador tab (\t)
    parts := strings.Split(line, "\t")
    if len(parts) < 6 {
        return "", "", "", "", "", "", fmt.Errorf("linha de log inválida")
    }

    // Parse do timestamp removendo a parte do fuso horário
    timestamp := parts[0]
    // Remove o sufixo de fuso horário, caso esteja presente
    timestamp = strings.TrimSuffix(timestamp, " +0000") // Remove o fuso horário

    // Tenta parse do timestamp para o formato correto
    parsedTime, err := time.Parse("2006-01-02 15:04:05", timestamp)
    if err != nil {
        return "", "", "", "", "", "", fmt.Errorf("erro ao parsear timestamp: %v", err)
    }

    // Retorna o timestamp no formato adequado para o MySQL
    formattedTimestamp := parsedTime.Format("2006-01-02 15:04:05")

    // O IP é o segundo campo
    ip := parts[1]

    // O username é o terceiro campo
    username := parts[2]

    // O interface é o quarto campo
    interfaceName := parts[3]

    // O clientIP é o quinto campo
    clientIP := parts[4]

    // A parte restante após o IP será a ação (em action)
    action := strings.Join(parts[5:], " ")

    // Retorna os dados processados
    return formattedTimestamp, ip, username, interfaceName, clientIP, action, nil
}

// Monitorar mudanças no arquivo de log
func monitorFile(filePath string, db *sql.DB) {
    // Cria um novo watcher para monitorar o arquivo
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    // Adicionar o arquivo para monitoramento
    err = watcher.Add(filePath)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Monitorando o arquivo:", filePath)

    // Processar os eventos do watcher
    for {
        select {
        case event := <-watcher.Events:
            // Se o arquivo foi escrito, processa o arquivo
            if event.Op&fsnotify.Write == fsnotify.Write {
                processFile(filePath, db)
            }
        case err := <-watcher.Errors:
            // Se ocorrer um erro, imprime o erro
            if err != nil {
                fmt.Println("Erro de monitoramento:", err)
            }
        }
    }
}

// Processar o arquivo de log
func processFile(filePath string, db *sql.DB) {
    // Abre o arquivo de log
    file, err := os.Open(filePath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Ler todas as linhas do arquivo
    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    // Verifica se há pelo menos uma linha
    if len(lines) > 0 {
        // Pega a última linha
        lastLine := lines[len(lines)-1]

        // Parse da última linha
        timestamp, ip, username, interfaceName, clientIP, action, err := parseLogLine(lastLine)
        if err != nil {
            fmt.Println("Erro ao processar linha do log:", err)
            return
        }

        // Inserir no banco de dados
        err = insertLogData(db, timestamp, ip, username, interfaceName, clientIP, action)
        if err != nil {
            fmt.Println("Erro ao inserir dados no banco:", err)
        }

        // Agora, reescreve o arquivo, excluindo a última linha
        newFile, err := os.Create(filePath)
        if err != nil {
            log.Fatal(err)
        }
        defer newFile.Close()

        // Escrever as linhas restantes (excluindo a última)
        for _, line := range lines[:len(lines)-1] {
            _, err := newFile.WriteString(line + "\n")
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}

func main() {
    // Conectar ao banco de dados MySQL
    db, err := connectDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Caminho do arquivo de log
    filePath := "/applog/tacacs.log"

    // Iniciar o monitoramento do arquivo de log
    monitorFile(filePath, db)
}
