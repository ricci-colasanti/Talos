# Talos
## Migration Microsimulation Engine

A fully self-contained, dynamically configurable demographic microsimulation system written in Go. **Talos is specifically designed as a migration microsimulation model** that simulates population movement between geographic areas while also handling other demographic processes like aging and mortality.

## Overview

Talos is a **migration-focused microsimulation engine** that models population dynamics through individual-level simulation. Unlike traditional microsimulation systems that require code changes for each new model, Talos uses:

- **SQL queries** for model logic (declarative and flexible)
- **YAML configuration** for model parameters, migration rates, and execution order
- **CSV files** for population data (works with any column structure)
- **Pure Go SQLite** for in-memory data processing (no external dependencies)

The result is a single, self-contained executable that can be distributed and run on any system without installation requirements. **Talos's primary strength is its ability to model complex migration patterns** with age-specific probabilities, area tracking, and configurable destination selection.

## Key Features

### Migration-Focused Capabilities
- **Age-Specific Migration Probabilities**: Different migration rates for children, young adults, middle-aged, and elderly populations
- **Area Tracking**: Tracks current and previous areas for migration history analysis
- **Random Destination Selection**: Migrants choose from multiple destination areas
- **Migration Statistics**: Automatic tracking of migration rates, flows, and patterns
- **Flexible Migration Logic**: Easily modify migration rules through YAML configuration

### General Features
- **Zero External Dependencies**: Pure Go implementation with embedded SQLite - no C compiler, no SQLite installation, no system libraries required
- **Single Binary Deployment**: Build once, run anywhere (Linux, Windows, macOS)
- **Fully Dynamic Data Handling**: Automatically detects and adapts to any CSV column structure
- **SQL-Based Models**: Define complex demographic transitions using familiar SQL syntax
- **Parameter Substitution**: Use placeholders like `{migration_rates.adult}` in SQL queries for flexible parameterization
- **Configurable Statistics**: Define any SQL query as a statistic to track population metrics
- **Priority-Based Execution**: Models run in specified priority order (age, mortality, migration)
- **In-Memory Processing**: Fast, in-memory SQLite database for population data
- **Checkpoint Output**: Results saved as CSV for further analysis
- **Reproducible Results**: Fixed random seeds for consistent simulation outcomes

## Migration Model

Talos's migration model is designed to be simple yet powerful, allowing researchers to simulate realistic population movement patterns.

### How Migration Works

1. **Current Area Tracking**: Each individual has an `area` column indicating their current location
2. **Previous Area Tracking**: The `previous_area` column stores where they were before migration
3. **Age-Based Probabilities**: Migration likelihood varies by age group (most mobile: 18-34; least: 65+)
4. **Random Destination**: When migrating, individuals move to a random area (configurable)
5. **Temporal Dynamics**: Migration probabilities apply annually during the simulation

### Migration Probabilities by Age

| Age Group | Probability (per year) | Rationale |
|-----------|----------------------|-----------|
| 0-17      | 2%                   | Children move with families |
| 18-34     | 8%                   | Most mobile - education, jobs, housing |
| 35-64     | 3%                   | Career-established, less mobile |
| 65+       | 1%                   | Retired, least mobile |

These rates are fully configurable in the YAML configuration file.

## Architecture

### Core Components

#### 1. **Model Loader and Executor**
- Parses YAML configuration files containing model definitions
- Dynamically extracts SQL queries and parameters from each model
- Substitutes parameter placeholders with actual values
- Executes SQL updates in priority order (age → mortality → migration)

#### 2. **Population Data Manager**
- Reads CSV files with automatic column detection
- Determines column types (int, bool, string) from data
- Creates SQLite tables dynamically with appropriate column types
- Handles missing or malformed data gracefully
- Supports area columns for migration tracking

#### 3. **Simulation Engine**
- Orchestrates the annual simulation loop
- Executes models in priority order
- Collects and displays migration statistics
- Saves checkpoint results
- Tracks area-level population changes

#### 4. **Migration Statistics Aggregator**
- Executes user-defined SQL migration statistics queries
- Tracks migration flows between areas
- Calculates migration rates and patterns
- Reports area-level population distributions
- Handles NULL values and type conversions

#### 5. **Parameter Substitution System**
- Replaces `{parameter}` placeholders in SQL queries
- Supports nested parameters like `{migration_rates.adult_18_34}`
- Handles multiple parameter types (int, float, string)
- Enables easy tuning of migration probabilities

### Data Flow

1. **Input**: CSV population file with area information is loaded
2. **Validation**: Column types are automatically detected (including area columns)
3. **Storage**: Data is loaded into in-memory SQLite database
4. **Simulation**: For each iteration (year):
   - Age model: Everyone ages by 1 year
   - Mortality model: Age-specific death probabilities applied
   - Migration model: Age-specific migration probabilities applied
   - Statistics: Migration patterns and population distributions are computed
5. **Output**: Final population state (including migration histories) is saved as CSV

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

Models define demographic transitions using SQL UPDATE statements. **Migration is a priority model** that runs after aging and mortality:

```yaml
models:
  - name: "age_increment"           # Age progression
    type: "sql_update"
    priority: 1
    enabled: true
    parameters:
      query: |
        UPDATE population 
        SET age = age + 1 
        WHERE alive = true

  - name: "mortality"               # Death model
    type: "sql_update"
    priority: 2
    enabled: true
    parameters:
      mortality_rates:
        infant: 0.005
        elderly: 0.10
      query: |
        UPDATE population 
        SET alive = false 
        WHERE alive = true AND ...
  
  - name: "migration"               # Migration model
    type: "sql_update"
    priority: 3
    enabled: true
    description: "Age-based migration between areas"
    parameters:
      migration_rates:
        child_0_17: 0.02
        adult_18_34: 0.08
        adult_35_64: 0.03
        elderly_65_plus: 0.01
      query: |
        UPDATE population 
        SET previous_area = area 
        WHERE alive = true;
        
        UPDATE population 
        SET area = 
          CASE 
            WHEN age < 18 AND alive = true AND random() < {migration_rates.child_0_17} 
              THEN CAST(abs(random() % 5) + 1 AS INTEGER)
            -- ... other age groups
            ELSE area
          END
        WHERE alive = true
```

### 3. Statistics Definitions

Statistics are SQL queries that produce summary metrics, including migration-specific statistics:

```yaml
statistics:
  - name: "migration_stats"
    description: "Migration statistics"
    query: |
      SELECT 
        COUNT(CASE WHEN previous_area != area THEN 1 END) as migrants,
        COUNT(*) as total,
        CAST(COUNT(CASE WHEN previous_area != area THEN 1 END) AS FLOAT) / COUNT(*) * 100 as migration_rate_pct
      FROM population
      WHERE alive = true
  
  - name: "area_distribution"
    description: "Population by area"
    query: |
      SELECT 
        area,
        COUNT(*) as total,
        COUNT(CASE WHEN alive = true THEN 1 END) as alive
      FROM population
      GROUP BY area
      ORDER BY area
```

## Library Dependencies

### External Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| `modernc.org/sqlite` | v1.53.0 | Pure Go SQLite driver with no CGO dependencies. Embeds SQLite in the Go binary, providing a full-featured SQL database without external installations. Supports transactions, indexes, and SQL functions like random(), CASE, and aggregate functions. Critical for migration queries with random destination selection. |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parser that converts configuration files into Go structs. Handles nested structures (like migration_rates), comments, and multi-line strings (like SQL queries). Uses reflection to map YAML fields to Go types. |

### Standard Library Packages

| Package | Purpose |
|---------|---------|
| `database/sql` | Standard Go SQL interface used for all database operations. Provides a clean abstraction layer that works with any SQL driver. Used for migration queries and statistics. |
| `encoding/csv` | Reads and writes CSV files. Handles quoting, escaping, and different delimiters. Used for both input population data and output results. |
| `fmt` | String formatting and printing. Used for log messages and error formatting. |
| `log` | Logging with timestamps. Provides simple, consistent logging across the application, including migration statistics output. |
| `math/rand` | Pseudo-random number generation. Provides random() function for probabilistic transitions including migration probability. Seeded for reproducibility of migration patterns. |
| `os` | Operating system functions for file I/O, command-line arguments, and environment variables. |
| `sort` | Sorting algorithms used to order models by priority (ensures migration runs after aging and mortality). |
| `strconv` | String conversion utilities. Parses integers from CSV data and converts numeric values to strings for output. |
| `strings` | String manipulation functions for trimming whitespace, splitting, joining, and replacing placeholders in SQL queries (e.g., substituting migration rates). |
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
The entry point. Parses command-line arguments, reads the YAML configuration, initializes the simulation, loads population data (including area information), executes the simulation loop (including migration), and saves results.

#### `loadPopulationDynamic(csvFile string) (*sql.DB, []ColumnInfo, error)`
**Purpose**: Loads population data from CSV with automatic column detection, including area columns.
**Parameters**:
- `csvFile`: Path to the CSV file
**Returns**:
- `*sql.DB`: SQLite database connection
- `[]ColumnInfo`: Detected column metadata (including area and previous_area)
- `error`: Any error encountered

**How it works**:
1. Reads the CSV header to discover column names (including area and previous_area)
2. Samples data rows to detect column types (int, bool, string)
3. Creates a SQLite table with appropriate column types
4. Inserts all data rows into the database
5. Handles missing or malformed data gracefully

#### `executeModel(db *sql.DB, model ModelConfig, verbose bool) error`
**Purpose**: Executes a single model's SQL query (including migration model).
**Parameters**:
- `db`: SQLite database connection
- `model`: Model configuration (could be age, mortality, or migration)
- `verbose`: Whether to log detailed output
**Returns**: Error if execution fails

**How it works**:
1. Extracts the SQL query from the model configuration
2. Substitutes parameter placeholders with actual values (e.g., migration rates)
3. Optionally logs the executed SQL
4. Executes the UPDATE statement against the population

#### `substituteParameters(query string, params map[string]interface{}) string`
**Purpose**: Replaces `{parameter}` placeholders with actual values.
**Parameters**:
- `query`: SQL query with placeholders (e.g., migration query with `{migration_rates.adult}`)
- `params`: Parameter map
**Returns**: SQL query with placeholders replaced

**Supports**:
- Simple parameters: `{rate}` → `0.01`
- Nested parameters: `{migration_rates.adult_18_34}` → `0.08`

#### `printStatistics(db *sql.DB, statistics []StatisticConfig, iteration int)`
**Purpose**: Executes and displays all configured statistics, including migration statistics.
**Parameters**:
- `db`: SQLite database connection
- `statistics`: List of statistic definitions (including migration_stats, area_distribution)
- `iteration`: Current simulation iteration number

**How it works**:
1. Executes each statistic's SQL query (including migration rate calculations)
2. Dynamically reads column names from results
3. Builds human-readable output with column names (e.g., "migrants: 2, migration_rate_pct: 10.53")
4. Handles NULL values and type conversions

#### `savePopulationDynamic(db *sql.DB, columns []ColumnInfo, outputFile string) error`
**Purpose**: Exports the population to CSV, including migration history.
**Parameters**:
- `db`: SQLite database connection
- `columns`: Column metadata (includes area and previous_area)
- `outputFile`: Output file path
**Returns**: Error if saving fails

**How it works**:
1. Queries all data from the population table (including area and previous_area)
2. Orders by person_id for consistency
3. Writes header and data rows to CSV
4. Handles type conversion for bool, int, float, string

### Helper Functions

- `filterEnabledModels(models []ModelConfig) []ModelConfig`: Filters to only enabled models
- `sortModelsByPriority(models []ModelConfig)`: Sorts models by priority (ensures migration runs after aging and mortality)
- `getColumnNames(columns []ColumnInfo) []string`: Extracts column names

### Data Structures

#### `ModelConfig`
```go
type ModelConfig struct {
    Name        string                 // Unique model identifier (e.g., "migration")
    Type        string                 // Model type (currently only "sql_update")
    Priority    int                    // Execution order (lower = earlier; migration typically priority 3)
    Enabled     bool                   // Whether model should run
    Description string                 // Human-readable description
    Parameters  map[string]interface{} // Model-specific parameters (e.g., migration_rates)
}
```

#### `StatisticConfig`
```go
type StatisticConfig struct {
    Name        string // Statistic identifier (e.g., "migration_stats")
    Description string // Human-readable description
    Query       string // SQL query to execute (e.g., migration rate calculation)
}
```

#### `SimulationConfig`
```go
type SimulationConfig struct {
    Simulation SimulationParameters
    Models     []ModelConfig       // Includes migration model
    Statistics []StatisticConfig   // Includes migration statistics
}
```

#### `ColumnInfo`
```go
type ColumnInfo struct {
    Name  string // Column name from CSV header (e.g., "area", "previous_area")
    Type  string // "int", "bool", or "string"
    IsKey bool   // Whether it's a primary key
}
```

## Usage Examples

### Basic Usage with Migration

```bash
# Build the binary
go build -o talos main.go

# Run with configuration (includes migration model)
./talos config.yaml

# Check migration statistics in output
grep "migration_stats" output.log
```

### Customizing Migration Rates

1. Edit `config.yaml`:
```yaml
parameters:
  migration_rates:
    child_0_17: 0.03      # Increase child migration
    adult_18_34: 0.12     # Increase young adult migration
    adult_35_64: 0.02     # Decrease middle-aged migration
    elderly_65_plus: 0.01 # Keep elderly migration low
```

2. No code changes required!

### Adding More Areas

Modify the random area generation in the migration query:
```sql
-- For 10 areas (1-10)
THEN CAST(abs(random() % 10) + 1 AS INTEGER)

-- For 3 areas with specific names
THEN CASE abs(random() % 3)
  WHEN 0 THEN 'London'
  WHEN 1 THEN 'Manchester'
  ELSE 'Birmingham'
END
```

### Tracking Migration by Age Group

Add this statistic to `config.yaml`:
```yaml
statistics:
  - name: "migration_by_age"
    description: "Migration by age group"
    query: |
      SELECT 
        CASE 
          WHEN age < 18 THEN '0-17'
          WHEN age < 35 THEN '18-34'
          WHEN age < 65 THEN '35-64'
          ELSE '65+'
        END as age_group,
        COUNT(CASE WHEN previous_area != area THEN 1 END) as migrants,
        COUNT(*) as total,
        CAST(COUNT(CASE WHEN previous_area != area THEN 1 END) AS FLOAT) / COUNT(*) * 100 as migration_rate_pct
      FROM population
      WHERE alive = true
      GROUP BY age_group
      ORDER BY age_group
```

## Input Data Format for Migration

The system accepts any CSV file with a header row. **For migration, the following columns are required or recommended:**

### Required Columns
- `person_id`: Unique identifier
- `age`: Age in years
- `area`: Current geographic area (integer or string)
- `alive`: Boolean indicating if individual is alive

### Recommended Columns
- `previous_area`: Previous geographic area (initialized to 0 or -1)

### Supported Column Types
- **Integer**: Any whole number (e.g., `25`, `5` for area IDs)
- **Boolean**: `true`, `false`, `True`, `False`, `1`, `0`
- **String**: Everything else (e.g., `London`, `Manchester` for area names)

### Example CSV with Migration Data
```csv
person_id,age,sex,area,alive,previous_area
1,25,F,1,true,0
2,30,M,2,true,0
3,45,F,1,true,0
4,68,M,3,true,0
5,82,F,2,true,0
6,2,M,4,true,0
```

## Performance Considerations

- **In-Memory Database**: All data is stored in memory for maximum speed
- **SQL Optimization**: SQLite optimizer handles query planning
- **Batch Processing**: INSERT and UPDATE operations are batched when possible
- **Indexing**: Primary keys are automatically indexed for fast lookups
- **Memory Usage**: Approximately 1-2MB per 1000 individuals
- **Migration Performance**: Migration runs as a single SQL update, making it very fast even for large populations

## Extending the Migration Model

### Adding New SQL Functions for Migration

The SQLite driver supports custom functions. You could extend the system with migration-specific functions:

```go
// Register a function to calculate migration distance
db.Exec("CREATE FUNCTION migration_distance(from_area INT, to_area INT) RETURNS INT AS '...'")
```

### Adding Destination Preferences

Modify the migration query to include destination preferences:
```sql
-- Weighted random: 30% chance of area 1, 20% chance of area 2
THEN CASE 
  WHEN random() < 0.3 THEN 1
  WHEN random() < 0.5 THEN 2
  ELSE CAST(abs(random() % 5) + 1 AS INTEGER)
END
```

### Adding Distance-Based Migration

```sql
-- Only move to adjacent areas
THEN area + (CASE 
  WHEN random() < 0.33 THEN -1
  WHEN random() < 0.66 THEN 1
  ELSE 0
END)
```

### Adding Household Migration

For household-level migration, you could:
1. Identify households (e.g., by family_id)
2. Move all household members together
3. Track household-level migration patterns

### Parallel Processing

For very large populations, you could add parallel processing by splitting the population by area and processing each area in a separate goroutine:

```go
areaGroups := splitByArea(population)
var wg sync.WaitGroup
for _, area := range areaGroups {
    wg.Add(1)
    go processAreaMigration(area, &wg)
}
wg.Wait()
```

## Troubleshooting

### Common Migration Issues

**"No migration occurring"**
- Check that migration model is enabled in config.yaml
- Verify migration probabilities are > 0
- Ensure area column exists in population data
- Check that individuals are alive before migration

**"All migrants go to same area"**
- Check the random area selection logic in the SQL query
- Ensure `abs(random() % N) + 1` is using the correct N for number of areas
- Verify the random seed isn't fixed to a single value

**"Migration rate too high/low"**
- Adjust migration_rates parameters in config.yaml
- Check that probabilities are per-year (should be < 0.1 for realistic rates)
- Verify the random() function is working correctly

**"Previous_area not updating"**
- Check the query includes `SET previous_area = area` before migration
- Verify the update runs before the migration update
- Ensure previous_area column exists in the population table

**"No such function: random"**
- Some SQLite builds may not include the random() function
- Use `abs(random() % 1000000) / 1000000.0` instead

### General Issues

**"Failed to load population"**
- Check CSV file path and format
- Ensure CSV has a header row
- Verify all rows have the same number of columns
- Check area column exists

**"Failed to execute model query"**
- Check SQL syntax in the model definition
- Verify column names exist in the population (e.g., area, previous_area)
- Ensure parameter placeholders are correct

**Binary won't run on target system**
- Build for the target platform: `GOOS=linux GOARCH=amd64 go build`
- Check architecture compatibility: `file talos`

## License

MIT License - See LICENSE file for details.

## Authors

Nik Lomax  
Alison Heppenstall  
Andreas Hoehn  
Hugh Rice  
Ric Colasanti  

## Acknowledgments

This system builds upon the excellent work of:
- **SQLite**: The world's most used database engine
- **modernc.org/sqlite**: Pure Go port of SQLite
- **Go**: The programming language that makes it all possible
- **Migration Research Community**: For insights into population movement patterns

---

**Talos: Bringing migration models to life through simple, self-contained, SQL-powered simulation.** 🏛️