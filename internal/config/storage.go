package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dancaldera/mirador/internal/models"
)

// GetConfigDir returns the configuration directory for the application
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".mirador")
	return configDir, os.MkdirAll(configDir, 0755)
}

// GetConnectionsFile returns the path to the connections file
func GetConnectionsFile() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "connections.json"), nil
}

// LoadSavedConnections loads saved connections from the configuration file
func LoadSavedConnections() ([]models.SavedConnection, error) {
	connectionsFile, err := GetConnectionsFile()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(connectionsFile); os.IsNotExist(err) {
		return []models.SavedConnection{}, nil
	}

	data, err := os.ReadFile(connectionsFile)
	if err != nil {
		return nil, err
	}

	var connections []models.SavedConnection
	if err := json.Unmarshal(data, &connections); err != nil {
		// If we can't parse the file, return empty slice instead of error
		// This allows for graceful recovery from corrupted config files
		return []models.SavedConnection{}, nil
	}

	return connections, nil
}

// SaveConnections saves connections to the configuration file
func SaveConnections(connections []models.SavedConnection) error {
	connectionsFile, err := GetConnectionsFile()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(connectionsFile), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(connectionsFile, data, 0644)
}

// GetQueryHistoryFile returns the path to the query history file
func GetQueryHistoryFile() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "query_history.json"), nil
}

// LoadQueryHistory loads query history from the configuration file
func LoadQueryHistory() ([]models.QueryHistoryEntry, error) {
	historyFile, err := GetQueryHistoryFile()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		return []models.QueryHistoryEntry{}, nil
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	var history []models.QueryHistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		// If we can't parse the file, return empty slice instead of error
		return []models.QueryHistoryEntry{}, nil
	}

	return history, nil
}

// SaveQueryHistory saves query history to the configuration file
func SaveQueryHistory(history []models.QueryHistoryEntry) error {
	historyFile, err := GetQueryHistoryFile()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(historyFile), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

// ExportToCSV exports data to CSV format
func ExportToCSV(columns []string, rows [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	header := ""
	for i, col := range columns {
		if i > 0 {
			header += ","
		}
		// Quote columns that contain commas or quotes
		if strings.Contains(col, ",") || strings.Contains(col, "\"") {
			col = fmt.Sprintf("\"%s\"", strings.ReplaceAll(col, "\"", "\"\""))
		}
		header += col
	}
	header += "\n"

	if _, err := file.WriteString(header); err != nil {
		return err
	}

	// Write data rows
	for _, row := range rows {
		line := ""
		for i, cell := range row {
			if i > 0 {
				line += ","
			}
			// Quote cells that contain commas or quotes
			if strings.Contains(cell, ",") || strings.Contains(cell, "\"") {
				cell = fmt.Sprintf("\"%s\"", strings.ReplaceAll(cell, "\"", "\"\""))
			}
			line += cell
		}
		line += "\n"

		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// ExportToJSON exports data to JSON format
func ExportToJSON(columns []string, rows [][]string, filename string) error {
	var jsonData []map[string]string

	for _, row := range rows {
		rowMap := make(map[string]string)
		for i, col := range columns {
			if i < len(row) {
				rowMap[col] = row[i]
			} else {
				rowMap[col] = ""
			}
		}
		jsonData = append(jsonData, rowMap)
	}

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GenerateExportFilename generates a filename for exported data
func GenerateExportFilename(tableName, format string) string {
	timestamp := time.Now().Format("20060102_150405")
	if tableName != "" {
		return fmt.Sprintf("%s_%s.%s", tableName, timestamp, format)
	}
	return fmt.Sprintf("query_result_%s.%s", timestamp, format)
}
