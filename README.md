<div align="left">
  <img src="img/Talos.png" alt="Talos Logo" width="300" height="300">
  <p><em>Talos — Migration Microsimulation Engine</em></p>
</div>

# Talos
## Migration Microsimulation Engine

A fully self-contained, dynamically configurable demographic microsimulation system written in Go. **Talos is specifically designed as a migration microsimulation model** that simulates population movement between geographic areas while also handling other demographic processes like aging and mortality.

The design and architecture of Talos is described in the following paper:

> **SQL as a Domain-Specific Language for Demographic Microsimulation**

This paper discusses how SQL's declarative nature makes it ideal for demographic modeling, the Lisp-SQL connection, and the practical implementation of migration microsimulation using pure Go SQLite.

**Read the paper:** [SQL as a Domain-Specific Language for Demographic Microsimulation](SQL_as_a_Domain_Specific_Language_for_Demographic_Migration_Microsimulation.md.pdf)

## Overview

Talos is a **migration-focused microsimulation engine** that models population dynamics through individual-level simulation. Unlike traditional microsimulation systems that require code changes for each new model, Talos uses:

- **SQL queries** for model logic (declarative and flexible)
- **YAML configuration** for model parameters, migration rates, and execution order
- **CSV files** for population data (works with any column structure)
- **Pure Go SQLite** for in-memory data processing (no external dependencies)

The result is a single, self-contained executable that can be distributed and run on any system without installation requirements. **Talos's primary strength is its ability to model complex migration patterns** with age-specific probabilities, area tracking, and configurable destination selection.

## Tutorials

Get started with Talos through our step-by-step tutorials. Each tutorial builds on the previous one, taking you from beginner to advanced user.

| Tutorial | Description | Level |
|----------|-------------|-------|
| [Tutorial 1: Building an Aging Model](tutorials/tutorial1_aging.md) | Learn the basics by creating a simple aging model. You'll create a population CSV, write your first configuration, run the simulation, and analyze the output. | ⭐ Beginner |

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

## Quick Start

### No Installation Required!

Talos is distributed as pre-built executables for multiple platforms. **You don't need to install Go or compile anything** to use Talos - just download the executable for your platform and run it.

If you're a developer who wants to modify the source code or compile for a different platform, compilation instructions are provided in the [Compiling from Source](#compiling-from-source) section.

### Prerequisites

- None! Just download the executable for your platform
- A text editor (VS Code, Sublime, Notepad++, etc.) for editing configuration files
- **Optional**: Go 1.21+ if you want to compile from source

### Important: Specify Your ID Column

Before running Talos, make sure your `config.yaml` includes an `id_column`:

```yaml
simulation:
  id_column: "person_id"  # Replace with your ID column name
```

This column must exist in your population CSV and contain unique values for each individual.

### Download Pre-Built Binary (Easiest)

Pre-built executables are included in the release package. Choose your platform:

| Platform | Binary Name |
|----------|-------------|
| Linux (x86_64) | `talos-linux-amd64` |
| Linux (ARM64) | `talos-linux-arm64` |
| Windows (x86_64) | `talos-windows-amd64.exe` |
| macOS (Intel) | `talos-darwin-amd64` |
| macOS (Apple Silicon) | `talos-darwin-arm64` |

```bash
# On Linux/macOS
chmod +x talos-*
./talos-linux-amd64 config.yaml

# On Windows
talos-windows-amd64.exe config.yaml
```

### Run with Sample Data

```bash
# Make sure you have config.yaml and population.csv in the same directory
# Then run Talos with your configuration
./talos-linux-amd64 config.yaml
```

## Configuration

The system is entirely configured through a single YAML file:

### Simulation Parameters

```yaml
simulation:
  iterations: 10                    # Number of years to simulate
  population_file: "population.csv" # Input CSV file
  output_file: "output.csv"         # Output CSV file
  random_seed: 42                   # Fixed random seed for reproducibility
  verbose: true                     # Detailed logging output
  id_column: "person_id"            # REQUIRED: Primary key column for ordering
```

### Model Definitions

Models define demographic transitions using SQL UPDATE statements:

```yaml
models:
  - name: "age_increment"
    type: "sql_update"
    priority: 1
    enabled: true
    parameters:
      query: |
        UPDATE population 
        SET age = age + 1 
        WHERE alive = true

  - name: "mortality"
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
        WHERE alive = true AND (
          (age < 1 AND random() < {mortality_rates.infant}) OR
          (age >= 85 AND random() < {mortality_rates.elderly})
        )
  
  - name: "migration"
    type: "sql_update"
    priority: 3
    enabled: true
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
            WHEN age >= 18 AND age < 35 AND alive = true AND random() < {migration_rates.adult_18_34} 
              THEN CAST(abs(random() % 5) + 1 AS INTEGER)
            WHEN age >= 35 AND age < 65 AND alive = true AND random() < {migration_rates.adult_35_64} 
              THEN CAST(abs(random() % 5) + 1 AS INTEGER)
            WHEN age >= 65 AND alive = true AND random() < {migration_rates.elderly_65_plus} 
              THEN CAST(abs(random() % 5) + 1 AS INTEGER)
            ELSE area
          END
        WHERE alive = true
```

### Statistics Definitions

Statistics are SQL queries that produce summary metrics:

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

## Input Data Format

The system accepts any CSV file with a header row. Column types are automatically detected:

### Required Columns
- `person_id`: Unique identifier (must match `id_column` in config)
- `age`: Age in years
- `area`: Current geographic area (integer or string)
- `alive`: Boolean indicating if individual is alive

### Recommended Columns
- `previous_area`: Previous geographic area (initialized to 0 or -1)

### ID Column

Talos requires you to specify an **ID column** in the configuration. This column:
- Must contain unique values for each individual
- Is used as the PRIMARY KEY in the database
- Determines the order of rows in the output CSV

**Why is this required?** Talos uses the ID column to ensure consistent ordering of output data, making it easier to compare results across different runs.

**Example:**
```yaml
simulation:
  id_column: "person_id"  # Must match a column in your CSV
```

If the ID column is not specified or doesn't exist in the CSV, Talos will exit with an error.

### Example CSV
```csv
person_id,age,sex,area,alive,previous_area
1,25,F,1,true,0
2,30,M,2,true,0
3,45,F,1,true,0
4,68,M,3,true,0
5,82,F,2,true,0
```

## Output

### Console Output
The simulation outputs real-time statistics including:
- Population counts (alive/dead)
- Age distribution
- Area distribution
- Migration statistics (migrants, migration rate)
- Average age by area

### CSV Output
The final population state is saved as CSV with all columns preserved, including:
- Original columns (person_id, age, sex, area, alive)
- Migration history (previous_area)
- Additional columns added by models

**Note:** The output CSV is always ordered by the ID column specified in the configuration.

## Compiling from Source

**Note:** This section is optional. Pre-built executables are included in the release package. You only need to compile from source if you want to modify the code or build for a platform not included.

### Prerequisites for Compilation

- Go 1.21 or later
- Git (optional)

### Step 1: Install Go (if not already installed)

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang-go
# Or install latest version
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**macOS:**
```bash
brew install go
```

**Windows:**
Download and install from https://go.dev/dl/

### Step 2: Clone or Create Project

```bash
# Clone the repository
git clone https://github.com/your-repo/talos.git
cd talos

# Or create a new project
mkdir talos
cd talos
```

### Step 3: Get Dependencies

```bash
# Initialize Go module
go mod init talos

# Get dependencies (pure-Go SQLite, no CGO needed)
go get modernc.org/sqlite
go get gopkg.in/yaml.v3

# Tidy dependencies
go mod tidy
```

### Step 4: Compile

**Compile for Current Platform:**
```bash
go build -o talos main.go
```

**Compile for Specific Platforms:**

| Target Platform | Command |
|-----------------|---------|
| Linux (x86_64) | `GOOS=linux GOARCH=amd64 go build -o talos-linux-amd64 main.go` |
| Linux (ARM64) | `GOOS=linux GOARCH=arm64 go build -o talos-linux-arm64 main.go` |
| Windows | `GOOS=windows GOARCH=amd64 go build -o talos-windows-amd64.exe main.go` |
| macOS (Intel) | `GOOS=darwin GOARCH=amd64 go build -o talos-darwin-amd64 main.go` |
| macOS (Apple Silicon) | `GOOS=darwin GOARCH=arm64 go build -o talos-darwin-arm64 main.go` |

**Compile for All Platforms at Once:**

Create a `Makefile`:

```makefile
.PHONY: all linux windows mac linux-arm mac-arm

all: linux windows mac linux-arm mac-arm

linux:
	GOOS=linux GOARCH=amd64 go build -o talos-linux-amd64 main.go

linux-arm:
	GOOS=linux GOARCH=arm64 go build -o talos-linux-arm64 main.go

windows:
	GOOS=windows GOARCH=amd64 go build -o talos-windows-amd64.exe main.go

mac:
	GOOS=darwin GOARCH=amd64 go build -o talos-darwin-amd64 main.go

mac-arm:
	GOOS=darwin GOARCH=arm64 go build -o talos-darwin-arm64 main.go

clean:
	rm -f talos-*
```

Then run:
```bash
make all
```

### Step 5: Verify Compilation

**Check Binary:**
```bash
# Check file type
file talos-linux-amd64

# Check for dynamic dependencies (should show no external libraries)
ldd talos-linux-amd64  # On Linux
otool -L talos-darwin-amd64  # On macOS

# Check binary size
ls -lh talos-*
```

**Test Run:**
```bash
./talos-linux-amd64 config.yaml
```

### Step 6: (Optional) Install System-Wide

**Linux/macOS:**
```bash
sudo cp talos-linux-amd64 /usr/local/bin/talos
sudo chmod +x /usr/local/bin/talos

# Now run from anywhere
talos config.yaml
```

**Windows:**
Add the directory containing `talos-windows-amd64.exe` to your PATH, or rename to `talos.exe` and place in a convenient location.

## Troubleshooting

### Common Issues

**"ERROR: id_column is required in simulation section of config.yaml"**
- Add `id_column: "your_id_column"` to the `simulation` section in config.yaml
- Make sure the column name matches exactly with your CSV header

**"ID column 'person_id' not found in CSV header"**
- Check that the column name in `id_column` matches your CSV header exactly
- Available columns are listed in the error message
- Example: If your CSV has `id` instead of `person_id`, use `id_column: "id"`

**"No such function: random"**
- Some SQLite builds may not include the random() function
- Use `abs(random() % 1000000) / 1000000.0` instead

**"Failed to load population"**
- Check CSV file path and format
- Ensure CSV has a header row
- Verify all rows have the same number of columns
- Make sure `id_column` matches a column in the CSV

**"Failed to execute model query"**
- Check SQL syntax in the model definition
- Verify column names exist in the population (e.g., area, previous_area)
- Ensure parameter placeholders are correct

**Binary won't run on target system**
- Build for the target platform: `GOOS=linux GOARCH=amd64 go build`
- Check architecture compatibility: `file talos`

**"go: command not found"**
- Go is not installed or not in your PATH
- See "Compiling from Source" section above

**"Permission denied" on Linux/macOS**
- Make the binary executable: `chmod +x talos-*`

## Performance Considerations

- **In-Memory Database**: All data is stored in memory for maximum speed
- **SQL Optimization**: SQLite optimizer handles query planning
- **Batch Processing**: INSERT and UPDATE operations are batched when possible
- **Indexing**: Primary keys are automatically indexed for fast lookups
- **Memory Usage**: Approximately 1-2MB per 1000 individuals
- **Migration Performance**: Migration runs as a single SQL update, making it very fast even for large populations

### Performance Benchmarks

| Population Size | Memory Usage | Time per Iteration |
|-----------------|--------------|-------------------|
| 1,000 | ~2 MB | < 0.1 seconds |
| 10,000 | ~20 MB | ~0.5 seconds |
| 100,000 | ~200 MB | ~3 seconds |
| 1,000,000 | ~2 GB | ~30 seconds |

*Benchmarks on Intel Core i7, 16GB RAM, SSD*

## Extending the Migration Model

### Adding New Migration Rules

1. Edit `config.yaml`:
```yaml
models:
  - name: "migration"
    type: "sql_update"
    priority: 3
    enabled: true
    parameters:
      migration_rates:
        child_0_17: 0.02
        adult_18_34: 0.08
        # Add new age group
        student_18_25: 0.10
      query: |
        UPDATE population 
        SET area = 
          CASE 
            WHEN age >= 18 AND age < 25 AND alive = true AND random() < {migration_rates.student_18_25} 
              THEN CAST(abs(random() % 5) + 1 AS INTEGER)
            -- ... existing age groups
            ELSE area
          END
        WHERE alive = true
```

2. No code changes required!

### Adding Destination Preferences

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

For household-level migration, add a household_id column and modify the migration query:

```sql
-- Identify households where the head moves
WITH moving_households AS (
  SELECT DISTINCT household_id
  FROM population
  WHERE alive = true 
    AND age BETWEEN 18 AND 34 
    AND random() < 0.08
)
-- Move all members of those households
UPDATE population 
SET area = new_area
FROM (SELECT household_id, CAST(abs(random() % 5) + 1 AS INTEGER) as new_area
      FROM moving_households) m
WHERE population.household_id = m.household_id
  AND alive = true
```

## License

MIT License - See LICENSE file for details.

## Authors

Nik Lomax  
Alison Heppenstall  
Hugh Rice  
Andreas Hoehn  
Ric Colasanti
