package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"gopkg.in/yaml.v3"
)

// ModelConfig represents a single model definition in YAML
type ModelConfig struct {
	Name        string                 `yaml:"name"`
	Type        string                 `yaml:"type"`
	Priority    int                    `yaml:"priority"`
	Enabled     bool                   `yaml:"enabled"`
	Description string                 `yaml:"description"`
	Parameters  map[string]interface{} `yaml:"parameters"`
}

// StatisticConfig defines a statistic to compute and display
type StatisticConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Query       string `yaml:"query"`
}

// SimulationConfig holds the full simulation configuration
type SimulationConfig struct {
	Simulation SimulationParameters `yaml:"simulation"`
	Models     []ModelConfig        `yaml:"models"`
	Statistics []StatisticConfig    `yaml:"statistics"`
}

// SimulationParameters holds simulation-level settings
type SimulationParameters struct {
	Iterations     int    `yaml:"iterations"`
	PopulationFile string `yaml:"population_file"`
	OutputFile     string `yaml:"output_file"`
	RandomSeed     int64  `yaml:"random_seed"`
	Verbose        bool   `yaml:"verbose"`
}

// ColumnInfo stores metadata about a column
type ColumnInfo struct {
	Name    string
	Type    string // "int", "string", "bool"
	IsKey   bool
}

func main() {
	// Parse command line
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config.yaml>")
	}
	configFile := os.Args[1]

	// 1. Read and parse YAML config
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var simConfig SimulationConfig
	if err := yaml.Unmarshal(configBytes, &simConfig); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// Set random seed if provided
	if simConfig.Simulation.RandomSeed > 0 {
		rand.Seed(simConfig.Simulation.RandomSeed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	log.Printf("Starting simulation")
	log.Printf("Iterations: %d", simConfig.Simulation.Iterations)
	log.Printf("Population file: %s", simConfig.Simulation.PopulationFile)
	log.Printf("Models loaded: %d", len(simConfig.Models))
	log.Printf("Statistics defined: %d", len(simConfig.Statistics))

	// 2. Load population data dynamically
	db, columns, err := loadPopulationDynamic(simConfig.Simulation.PopulationFile)
	if err != nil {
		log.Fatalf("Failed to load population: %v", err)
	}
	defer db.Close()

	log.Printf("Loaded population with columns: %v", getColumnNames(columns))

	// 3. Filter enabled models and sort by priority
	enabledModels := filterEnabledModels(simConfig.Models)
	sortModelsByPriority(enabledModels)

	log.Printf("Enabled models: %d", len(enabledModels))
	for _, model := range enabledModels {
		log.Printf("  - %s (priority: %d)", model.Name, model.Priority)
	}

	// 4. Run the simulation
	for i := 0; i < simConfig.Simulation.Iterations; i++ {
		log.Printf("\n=== Iteration %d/%d ===", i+1, simConfig.Simulation.Iterations)

		// Execute each model in priority order
		for _, model := range enabledModels {
			if err := executeModel(db, model, simConfig.Simulation.Verbose); err != nil {
				log.Fatalf("Model '%s' execution failed: %v", model.Name, err)
			}
		}

		// Print statistics dynamically from config
		printStatistics(db, simConfig.Statistics, i+1)
	}

	// 5. Save final population dynamically
	if err := savePopulationDynamic(db, columns, simConfig.Simulation.OutputFile); err != nil {
		log.Fatalf("Failed to save population: %v", err)
	}

	log.Printf("\n=== Simulation Complete ===")
	log.Printf("Results saved to %s", simConfig.Simulation.OutputFile)
}

// loadPopulationDynamic loads CSV with automatic column detection
func loadPopulationDynamic(csvFile string) (*sql.DB, []ColumnInfo, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	file, err := os.Open(csvFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("CSV file is empty")
	}

	// Detect columns from header
	header := records[0]
	columns := make([]ColumnInfo, len(header))
	for i, col := range header {
		col = strings.TrimSpace(col)
		columns[i] = ColumnInfo{
			Name:  col,
			Type:  "string",
			IsKey: col == "id" || col == "person_id",
		}
	}

	// Determine column types from data
	if len(records) > 1 {
		for i := range columns {
			for j := 1; j < len(records) && j < 5; j++ {
				if j < len(records) && i < len(records[j]) {
					val := strings.TrimSpace(records[j][i])
					if val == "" {
						continue
					}
					if _, err := strconv.Atoi(val); err == nil {
						columns[i].Type = "int"
					} else if val == "true" || val == "false" || val == "True" || val == "False" || val == "1" || val == "0" {
						columns[i].Type = "bool"
					} else {
						columns[i].Type = "string"
					}
					break
				}
			}
		}
	}

	// Create table dynamically
	createSQL := "CREATE TABLE population ("
	for i, col := range columns {
		colType := "TEXT"
		switch col.Type {
		case "int":
			colType = "INTEGER"
		case "bool":
			colType = "BOOLEAN"
		default:
			colType = "TEXT"
		}
		createSQL += fmt.Sprintf("%s %s", col.Name, colType)
		if col.IsKey {
			createSQL += " PRIMARY KEY"
		}
		if i < len(columns)-1 {
			createSQL += ", "
		}
	}
	createSQL += ");"

	if _, err := db.Exec(createSQL); err != nil {
		return nil, nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Insert data dynamically
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	colNames := getColumnNames(columns)
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO population (%s) VALUES (%s)",
		strings.Join(colNames, ", "),
		strings.Join(placeholders, ", "))

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < len(columns) {
			log.Printf("Warning: Row %d has insufficient fields, skipping", i)
			continue
		}

		args := make([]interface{}, len(columns))
		for j := range columns {
			val := strings.TrimSpace(record[j])
			switch columns[j].Type {
			case "int":
				if val == "" {
					args[j] = 0
				} else {
					intVal, _ := strconv.Atoi(val)
					args[j] = intVal
				}
			case "bool":
				if val == "" {
					args[j] = false
				} else {
					args[j] = val == "true" || val == "True" || val == "1"
				}
			default:
				args[j] = val
			}
		}

		if _, err := stmt.Exec(args...); err != nil {
			return nil, nil, fmt.Errorf("failed to insert row %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Loaded %d individuals with %d columns", len(records)-1, len(columns))
	return db, columns, nil
}

// executeModel runs the SQL query defined in the model
func executeModel(db *sql.DB, model ModelConfig, verbose bool) error {
	queryInterface, ok := model.Parameters["query"]
	if !ok {
		return fmt.Errorf("model '%s' missing 'query' parameter", model.Name)
	}

	query, ok := queryInterface.(string)
	if !ok {
		return fmt.Errorf("model '%s' query is not a string", model.Name)
	}

	substitutedQuery := substituteParameters(query, model.Parameters)

	if verbose {
		log.Printf("  Executing model: %s (priority: %d)", model.Name, model.Priority)
		log.Printf("  Query: %s", substitutedQuery)
	} else {
		log.Printf("  Executing model: %s", model.Name)
	}

	_, err := db.Exec(substitutedQuery)
	if err != nil {
		return fmt.Errorf("failed to execute model query: %w", err)
	}

	return nil
}

// substituteParameters replaces {parameter} placeholders with values
func substituteParameters(query string, params map[string]interface{}) string {
	result := query

	for key, value := range params {
		if key == "query" {
			continue
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			for nestedKey, nestedValue := range nestedMap {
				placeholder := fmt.Sprintf("{%s.%s}", key, nestedKey)
				result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", nestedValue))
			}
		} else {
			placeholder := fmt.Sprintf("{%s}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return result
}

// printStatistics executes and displays all statistics from config
func printStatistics(db *sql.DB, statistics []StatisticConfig, iteration int) {
	if len(statistics) == 0 {
		return
	}

	log.Printf("  Statistics:")

	for _, stat := range statistics {
		query := stat.Query

		// Execute the query
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("    ⚠️  Failed to compute '%s': %v", stat.Name, err)
			continue
		}
		defer rows.Close()

		// Get column names from the result
		colNames, err := rows.Columns()
		if err != nil {
			log.Printf("    ⚠️  Failed to get columns for '%s': %v", stat.Name, err)
			continue
		}

		// Check if we have results
		if !rows.Next() {
			log.Printf("    %s: No results", stat.Name)
			continue
		}

		// Create scan targets dynamically
		scanTargets := make([]interface{}, len(colNames))
		for i := range colNames {
			var val interface{}
			scanTargets[i] = &val
		}

		if err := rows.Scan(scanTargets...); err != nil {
			log.Printf("    ⚠️  Failed to scan '%s': %v", stat.Name, err)
			continue
		}

		// Build result string
		resultParts := make([]string, len(colNames))
		for i, colName := range colNames {
			val := *(scanTargets[i].(*interface{}))
			if val == nil {
				resultParts[i] = fmt.Sprintf("%s: NULL", colName)
			} else {
				switch v := val.(type) {
				case int64:
					resultParts[i] = fmt.Sprintf("%s: %d", colName, v)
				case float64:
					if v == float64(int64(v)) {
						resultParts[i] = fmt.Sprintf("%s: %.0f", colName, v)
					} else {
						resultParts[i] = fmt.Sprintf("%s: %.2f", colName, v)
					}
				case string:
					resultParts[i] = fmt.Sprintf("%s: %s", colName, v)
				case bool:
					resultParts[i] = fmt.Sprintf("%s: %v", colName, v)
				default:
					resultParts[i] = fmt.Sprintf("%s: %v", colName, v)
				}
			}
		}

		description := ""
		if stat.Description != "" {
			description = fmt.Sprintf(" (%s)", stat.Description)
		}
		log.Printf("    %s%s: %s", stat.Name, description, strings.Join(resultParts, ", "))
	}
}

// savePopulationDynamic exports the final population to CSV dynamically
func savePopulationDynamic(db *sql.DB, columns []ColumnInfo, outputFile string) error {
	colNames := getColumnNames(columns)

	query := fmt.Sprintf("SELECT %s FROM population ORDER BY person_id", strings.Join(colNames, ", "))

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query population: %w", err)
	}
	defer rows.Close()

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write(colNames)

	// Write data
	for rows.Next() {
		scanTargets := make([]interface{}, len(columns))
		for i := range columns {
			var val interface{}
			scanTargets[i] = &val
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		record := make([]string, len(columns))
		for i, target := range scanTargets {
			val := *(target.(*interface{}))
			if val == nil {
				record[i] = ""
			} else {
				switch v := val.(type) {
				case bool:
					if v {
						record[i] = "true"
					} else {
						record[i] = "false"
					}
				case int64:
					record[i] = strconv.FormatInt(v, 10)
				case float64:
					record[i] = strconv.FormatFloat(v, 'f', -1, 64)
				case string:
					record[i] = v
				default:
					record[i] = fmt.Sprintf("%v", v)
				}
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// getColumnNames returns a slice of column names
func getColumnNames(columns []ColumnInfo) []string {
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = col.Name
	}
	return names
}

// filterEnabledModels returns only enabled models
func filterEnabledModels(models []ModelConfig) []ModelConfig {
	var enabled []ModelConfig
	for _, model := range models {
		if model.Enabled {
			enabled = append(enabled, model)
		}
	}
	return enabled
}

// sortModelsByPriority sorts models by priority (lower = higher priority)
func sortModelsByPriority(models []ModelConfig) {
	sort.Slice(models, func(i, j int) bool {
		return models[i].Priority < models[j].Priority
	})
}
