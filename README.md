# Talos
# Microsimulation Engine

A fully self-contained, dynamically configurable demographic microsimulation system written in Go. This engine enables demographic researchers and analysts to define population models and statistics through simple YAML configuration files, without writing any code.

## Overview

This microsimulation engine simulates demographic processes (aging, mortality, education, income, migration, etc.) on individual-level population data. Unlike traditional microsimulation systems that require code changes for each new model, this system uses:

- **SQL queries** for model logic (declarative and flexible)
- **YAML configuration** for model parameters and execution order
- **CSV files** for population data (works with any column structure)
- **Pure Go SQLite** for in-memory data processing (no external dependencies)

The result is a single, self-contained executable that can be distributed and run on any system without installation requirements.

## Key Features

- **Zero External Dependencies**: Pure Go implementation with embedded SQLite - no C compiler, no SQLite installation, no system libraries required
- **Single Binary Deployment**: Build once, run anywhere (Linux, Windows, macOS)
- **Fully Dynamic Data Handling**: Automatically detects and adapts to any CSV column structure
- **SQL-Based Models**: Define complex demographic transitions using familiar SQL syntax
- **Parameter Substitution**: Use placeholders like `{mortality_rates.infant}` in SQL queries for flexible parameterization
- **Configurable Statistics**: Define any SQL query as a statistic to track population metrics
- **Priority-Based Execution**: Models run in specified priority order
- **In-Memory Processing**: Fast, in-memory SQLite database for population data
- **Checkpoint Output**: Results saved as CSV for further analysis

## Architecture

### Core Components

#### 1. **Model Loader and Executor**
- Parses YAML configuration files containing model definitions
- Dynamically extracts SQL queries and parameters from each model
- Substitutes parameter placeholders with actual values
- Executes SQL updates in priority order

#### 2. **Population Data Manager**
- Reads CSV files with automatic column detection
- Determines column types (int, bool, string) from data
- Creates SQLite tables dynamically with appropriate column types
- Handles missing or malformed data gracefully

#### 3. **Simulation Engine**
- Orchestrates the annual simulation loop
- Executes models in priority order
- Collects and displays statistics
- Saves checkpoint results

#### 4. **Statistics Aggregator**
- Executes user-defined SQL statistics queries
- Dynamically displays results with column names
- Handles NULL values and type conversions
- Reports results in human-readable format

#### 5. **Parameter Substitution System**
- Replaces `{parameter}` placeholders in SQL queries
- Supports nested parameters like `{mortality_rates.infant}`
- Handles multiple parameter types (int, float, string)

### Data Flow

1. **Input**: CSV population file is loaded
2. **Validation**: Column types are automatically detected
3. **Storage**: Data is loaded into in-memory SQLite database
4. **Simulation**: For each iteration (year):
   - Models are executed in priority order
   - Each model applies SQL updates to the population
   - Statistics are computed and displayed
5. **Output**: Final population state is saved as CSV

## Configuration File (`config.yaml`)

The system is entirely configured through a single YAML file with three main sections:

### 1. Simulation Parameters

```yaml
simulation:
  iterations: 10                    # Number of years to simulate
  population_file: "population.csv" # Input CSV file
  output_file: "output.csv"         # Output CSV file
  random_seed: 42                   # Fixed random seed for reproducibility
  verbose: true                     # Detailed logging output
```

### 2. Model Definitions

Models define demographic transitions using SQL UPDATE statements:

```yaml
models:
  - name: "age_increment"           # Model identifier
    type: "sql_update"              # Model type (currently only sql_update)
    priority: 1                     # Execution order (lower = earlier)
    enabled: true                   # Enable/disable without removing
    description: "Age progression"  # Human-readable description
    parameters:                     # Model parameters
      query: |                      # SQL UPDATE statement
        UPDATE population 
        SET age = age + 1 
        WHERE alive = true
```

**Parameter Substitution Example:**
```yaml
parameters:
  mortality_rates:
    infant: 0.005
    elderly: 0.10
  query: |
    UPDATE population 
    SET alive = false 
    WHERE age < 1 AND random() < {mortality_rates.infant}
```

### 3. Statistics Definitions

Statistics are SQL queries that produce summary metrics:

```yaml
statistics:
  - name: "population_total"
    description: "Total population"
    query: "SELECT COUNT(*) as total FROM population"
  
  - name: "age_distribution"
    description: "Age groups"
    query: |
      SELECT 
        COUNT(CASE WHEN age < 18 THEN 1 END) as children,
        COUNT(CASE WHEN age >= 18 AND age < 65 THEN 1 END) as adults,
        COUNT(CASE WHEN age >= 65 THEN 1 END) as elderly
      FROM population
```

## Library Dependencies

### External Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| `modernc.org/sqlite` | v1.53.0 | Pure Go SQLite driver with no CGO dependencies. Embeds SQLite in the Go binary, providing a full-featured SQL database without external installations. Supports transactions, indexes, and SQL functions like random(), CASE, and aggregate functions. |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parser that converts configuration files into Go structs. Handles nested structures, comments, and multi-line strings. Uses reflection to map YAML fields to Go types. |

### Standard Library Packages

| Package | Purpose |
|---------|---------|
| `database/sql` | Standard Go SQL interface used for all database operations. Provides a clean abstraction layer that works with any SQL driver. |
| `encoding/csv` | Reads and writes CSV files. Handles quoting, escaping, and different delimiters. Used for both input population data and output results. |
| `fmt` | String formatting and printing. Used for log messages and error formatting. |
| `log` | Logging with timestamps. Provides simple, consistent logging across the application. |
| `math/rand` | Pseudo-random number generation. Provides random() function for probabilistic transitions. Seeded for reproducibility. |
| `os` | Operating system functions for file I/O, command-line arguments, and environment variables. |
| `sort` | Sorting algorithms used to order models by priority. Implements efficient quicksort for slices. |
| `strconv` | String conversion utilities. Parses integers from CSV data and converts numeric values to strings for output. |
| `strings` | String manipulation functions for trimming whitespace, splitting, joining, and replacing placeholders in SQL queries. |
| `time` | Time functions used for generating random seeds and optional timestamp logging. |

### Indirect Dependencies (Transitive)

These libraries are automatically included via `modernc.org/sqlite`:

| Library | Purpose |
|---------|---------|
| `modernc.org/libc` | Pure Go implementation of standard C library functions needed by the SQLite port. Provides memory management and system calls without CGO. |
| `modernc.org/mathutil` | Mathematical utilities used by the SQLite implementation. |
| `modernc.org/memory` | Memory management utilities for the pure Go SQLite driver. |
| `golang.org/x/sys` | Low-level system call interfaces for cross-platform support. Handles OS-specific operations. |
| `github.com/dustin/go-humanize` | Human-readable formatting (used internally by SQLite driver). |
| `github.com/google/uuid` | UUID generation (used internally by SQLite driver). |
| `github.com/mattn/go-isatty` | Terminal detection for colored output (used internally). |
| `github.com/ncruces/go-strftime` | Time formatting utilities (used internally). |
| `github.com/remyoudompheng/bigfft` | Fast Fourier Transform for big integers (used internally). |

## Code Structure

### Main Functions

#### `main()`
The entry point. Parses command-line arguments, reads the YAML configuration, initializes the simulation, loads population data, executes the simulation loop, and saves results.

#### `loadPopulationDynamic(csvFile string) (*sql.DB, []ColumnInfo, error)`
**Purpose**: Loads population data from CSV with automatic column detection.
**Parameters**:
- `csvFile`: Path to the CSV file
**Returns**:
- `*sql.DB`: SQLite database connection
- `[]ColumnInfo`: Detected column metadata
- `error`: Any error encountered

**How it works**:
1. Reads the CSV header to discover column names
2. Samples data rows to detect column types (int, bool, string)
3. Creates a SQLite table with appropriate column types
4. Inserts all data rows into the database
5. Handles missing or malformed data gracefully

#### `executeModel(db *sql.DB, model ModelConfig, verbose bool) error`
**Purpose**: Executes a single model's SQL query.
**Parameters**:
- `db`: SQLite database connection
- `model`: Model configuration
- `verbose`: Whether to log detailed output
**Returns**: Error if execution fails

**How it works**:
1. Extracts the SQL query from the model configuration
2. Substitutes parameter placeholders with actual values
3. Optionally logs the executed SQL
4. Executes the UPDATE statement against the population

#### `substituteParameters(query string, params map[string]interface{}) string`
**Purpose**: Replaces `{parameter}` placeholders with actual values.
**Parameters**:
- `query`: SQL query with placeholders
- `params`: Parameter map
**Returns**: SQL query with placeholders replaced

**Supports**:
- Simple parameters: `{rate}` → `0.01`
- Nested parameters: `{mortality_rates.infant}` → `0.005`

#### `printStatistics(db *sql.DB, statistics []StatisticConfig, iteration int)`
**Purpose**: Executes and displays all configured statistics.
**Parameters**:
- `db`: SQLite database connection
- `statistics`: List of statistic definitions
- `iteration`: Current simulation iteration number

**How it works**:
1. Executes each statistic's SQL query
2. Dynamically reads column names from results
3. Builds human-readable output with column names
4. Handles NULL values and type conversions

#### `savePopulationDynamic(db *sql.DB, columns []ColumnInfo, outputFile string) error`
**Purpose**: Exports the population to CSV.
**Parameters**:
- `db`: SQLite database connection
- `columns`: Column metadata
- `outputFile`: Output file path
**Returns**: Error if saving fails

**How it works**:
1. Queries all data from the population table
2. Orders by person_id for consistency
3. Writes header and data rows to CSV
4. Handles type conversion for bool, int, float, string

#### Helper Functions

- `filterEnabledModels(models []ModelConfig) []ModelConfig`: Filters to only enabled models
- `sortModelsByPriority(models []ModelConfig)`: Sorts models by priority (lower first)
- `getColumnNames(columns []ColumnInfo) []string`: Extracts column names

### Data Structures

#### `ModelConfig`
```go
type ModelConfig struct {
    Name        string                 // Unique model identifier
    Type        string                 // Model type (currently only "sql_update")
    Priority    int                    // Execution order (lower = earlier)
    Enabled     bool                   // Whether model should run
    Description string                 // Human-readable description
    Parameters  map[string]interface{} // Model-specific parameters
}
```

#### `StatisticConfig`
```go
type StatisticConfig struct {
    Name        string // Statistic identifier
    Description string // Human-readable description
    Query       string // SQL query to execute
}
```

#### `SimulationConfig`
```go
type SimulationConfig struct {
    Simulation SimulationParameters
    Models     []ModelConfig
    Statistics []StatisticConfig
}
```

#### `ColumnInfo`
```go
type ColumnInfo struct {
    Name  string // Column name from CSV header
    Type  string // "int", "bool", or "string"
    IsKey bool   // Whether it's a primary key
}
```

## Usage Examples

### Basic Usage

```bash
# Build the binary
go build -o microsim main.go

# Run with configuration
./microsim config.yaml

# Run with different config
./microsim my_alternative_config.yaml
```

### Adding a New Model

1. Edit `config.yaml`:
```yaml
models:
  - name: "fertility"
    type: "sql_update"
    priority: 3
    enabled: true
    description: "Birth model"
    parameters:
      fertility_rates:
        age_20_29: 0.12
        age_30_39: 0.08
      query: |
        INSERT INTO population (age, sex, area, alive)
        SELECT 
          0 as age,
          'F' as sex,
          area,
          true as alive
        FROM population 
        WHERE sex = 'F' 
          AND age BETWEEN 20 AND 39
          AND alive = true
          AND (
            (age BETWEEN 20 AND 29 AND random() < {fertility_rates.age_20_29}) OR
            (age BETWEEN 30 AND 39 AND random() < {fertility_rates.age_30_39})
          )
```

2. No code changes required!

### Adding a New Statistic

```yaml
statistics:
  - name: "dependency_ratio"
    description: "Age dependency ratio"
    query: |
      SELECT 
        CAST(COUNT(CASE WHEN age < 15 OR age >= 65 THEN 1 END) AS FLOAT) /
        CAST(COUNT(CASE WHEN age >= 15 AND age < 65 THEN 1 END) AS FLOAT) 
        AS dependency_ratio
      FROM population
      WHERE alive = true
```

## Input Data Format

The system accepts any CSV file with a header row. Column types are automatically detected:

### Supported Column Types
- **Integer**: Any whole number (e.g., `25`, `-5`)
- **Boolean**: `true`, `false`, `True`, `False`, `1`, `0`
- **String**: Everything else (e.g., `M`, `Female`, `tertiary`)

### Example CSV
```csv
person_id,age,sex,area,alive,education_level,income
1,25,F,1,true,tertiary,45000
2,30,M,1,true,secondary,35000
3,45,F,1,true,tertiary,52000
```

## Performance Considerations

- **In-Memory Database**: All data is stored in memory for maximum speed
- **SQL Optimization**: SQLite optimizer handles query planning
- **Batch Processing**: INSERT and UPDATE operations are batched when possible
- **Indexing**: Primary keys are automatically indexed for fast lookups
- **Memory Usage**: Approximately 1-2MB per 1000 individuals

## Extending the System

### Adding New SQL Functions

The SQLite driver supports custom functions. You could extend the system by registering Go functions as SQL functions:

```go
db := sql.Open("sqlite", "file::memory:?cache=shared")
// Register custom SQL function
db.Exec("CREATE FUNCTION mortality_rate(age INT, sex TEXT) RETURNS FLOAT AS '...'")
```

### Adding New Model Types

Currently, only "sql_update" models are supported. You could extend `executeModel()` to support additional types:

```go
switch model.Type {
case "sql_update":
    // Current implementation
case "custom_model":
    // Your custom model logic
case "go_function":
    // Call a Go function
}
```

### Parallel Processing

For very large populations, you could add parallel processing by splitting the population by area and processing each area in a separate goroutine:

```go
areaGroups := splitByArea(population)
var wg sync.WaitGroup
for _, area := range areaGroups {
    wg.Add(1)
    go processArea(area, &wg)
}
wg.Wait()
```

## Troubleshooting

### Common Issues

**"No such function: random"**
- Some SQLite builds may not include the random() function
- Use `abs(random() % 1000000) / 1000000.0` instead

**"Failed to load population"**
- Check CSV file path and format
- Ensure CSV has a header row
- Verify all rows have the same number of columns

**"Failed to execute model query"**
- Check SQL syntax in the model definition
- Verify column names exist in the population
- Ensure parameter placeholders are correct

**Binary won't run on target system**
- Build for the target platform: `GOOS=linux GOARCH=amd64 go build`
- Check architecture compatibility: `file microsim`

## License

MIT License - See LICENSE file for details.

## Authors

Nik Lomax Alison Heppenstall Andreas Hoehn Hugh Rice Ric Colasanti

## Acknowledgments

This system builds upon the excellent work of:
- **SQLite**: The world's most used database engine
- **modernc.org/sqlite**: Pure Go port of SQLite
- **Go**: The programming language that makes it all possible
