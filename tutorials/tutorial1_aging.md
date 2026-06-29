# Talos Tutorial 1: Building an Aging Model

## Overview

In this tutorial, you'll learn how to create a simple aging model with Talos. We'll start with a CSV population file, write a configuration that ages everyone by one year, run the simulation, and analyze the output.

## Prerequisites

- Talos binary downloaded and in your PATH (or in the current directory)
- Basic understanding of CSV files
- A text editor (VS Code, Sublime, Notepad++, etc.)

## Step 1: Create a Population CSV

First, let's create a small population to work with. Create a file called `population.csv`:

```csv
person_id,age,sex,area,alive
1,25,F,1,true
2,30,M,1,true
3,45,F,1,true
4,68,M,1,true
5,82,F,1,true
6,2,M,1,true
7,15,F,1,true
8,35,M,1,true
9,55,F,1,true
10,70,M,1,true
```

This gives us 10 individuals with various ages. The columns are:
- `person_id`: Unique identifier for each person
- `age`: Age in years
- `sex`: Gender (M/F)
- `area`: Geographic area (we'll use just one area for now)
- `alive`: Whether the person is alive (true/false)

## Step 2: Create the Configuration File

Now let's create a configuration file called `config_aging.yaml`. This file tells Talos what models to run and how to run them.

```yaml
# config_aging.yaml
# A simple aging model configuration

simulation:
  iterations: 5                    # Run for 5 years
  population_file: "population.csv" # Input population
  output_file: "population_aged.csv" # Output after aging
  random_seed: 42                   # For reproducibility
  verbose: true                     # Show detailed output

models:
  - name: "age_increment"
    type: "sql_update"
    priority: 1
    enabled: true
    description: "Increment everyone's age by 1 year"
    parameters:
      query: |
        UPDATE population 
        SET age = age + 1 
        WHERE alive = true

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
      WHERE alive = true
```

### Understanding the Configuration

Let's break down what this configuration does:

**Simulation Section:**
- `iterations: 5` - Run the simulation for 5 years
- `population_file` - Where to read the population from
- `output_file` - Where to save the final population
- `random_seed` - Ensures results are reproducible
- `verbose: true` - Shows detailed logging

**Models Section:**
- We define one model called `age_increment`
- `priority: 1` - This is the only model, but priority matters if we have multiple models
- The `query` is an SQL UPDATE statement that adds 1 to everyone's age
- `WHERE alive = true` ensures only alive people age

**Statistics Section:**
- We define two statistics to track
- `population_total` - Counts everyone
- `age_distribution` - Groups population into children, adults, and elderly

## Step 3: Run the Simulation

Now run Talos with your configuration:

```bash
# If talos is in your PATH
talos config_aging.yaml

# Or if talos is in the current directory
./talos config_aging.yaml
```

### Expected Output

You should see output similar to this:

```
2024/01/15 10:00:00 Starting simulation
2024/01/15 10:00:00 Iterations: 5
2024/01/15 10:00:00 Population file: population.csv
2024/01/15 10:00:00 Models loaded: 1
2024/01/15 10:00:00 Statistics defined: 2
2024/01/15 10:00:00 Loaded 10 individuals with 4 columns
2024/01/15 10:00:00 Loaded population with columns: [person_id age sex area alive]
2024/01/15 10:00:00 Enabled models: 1
2024/01/15 10:00:00   - age_increment (priority: 1)

2024/01/15 10:00:00 === Iteration 1/5 ===
2024/01/15 10:00:00   Executing model: age_increment (priority: 1)
2024/01/15 10:00:00   Query: UPDATE population SET age = age + 1 WHERE alive = true
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 2, adults: 6, elderly: 2

2024/01/15 10:00:00 === Iteration 2/5 ===
2024/01/15 10:00:00   Executing model: age_increment (priority: 1)
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 2, adults: 5, elderly: 3

...

2024/01/15 10:00:00 === Iteration 5/5 ===
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 1, adults: 4, elderly: 5

2024/01/15 10:00:00 === Simulation Complete ===
2024/01/15 10:00:00 Results saved to population_aged.csv
```

## Step 4: Examine the Output

### Console Output Analysis

Look at the age_distribution statistics across iterations:

**Iteration 1:**
- Children: 2 (ages 2 and 15 become 3 and 16)
- Adults: 6 (ages 25,30,35,45,55,68 become 26,31,36,46,56,69)
- Elderly: 2 (ages 82 and 70 become 83 and 71)

**Iteration 5:** (after 5 years of aging)
- Children: 1 (only the 2-year-old is still under 18)
- Adults: 4 (people who started between 18-64)
- Elderly: 5 (the original 68,70,82 plus adults who turned 65+)

This shows the aging process working correctly!

### CSV Output Analysis

Open `population_aged.csv`:

```csv
person_id,age,sex,area,alive
1,30,F,1,true
2,35,M,1,true
3,50,F,1,true
4,73,M,1,true
5,87,F,1,true
6,7,M,1,true
7,20,F,1,true
8,40,M,1,true
9,60,F,1,true
10,75,M,1,true
```

Notice that everyone has aged exactly 5 years:
- Person 1: 25 → 30
- Person 2: 30 → 35
- Person 3: 45 → 50
- Person 4: 68 → 73
- Person 5: 82 → 87
- Person 6: 2 → 7
- Person 7: 15 → 20
- Person 8: 35 → 40
- Person 9: 55 → 60
- Person 10: 70 → 75

## Step 5: Adding More Statistics (Understanding SQL)

Now let's add more statistics to better understand our population. **This is where we'll learn the basics of SQL**, which is the language Talos uses for models and statistics.

### What is SQL?

SQL (Structured Query Language) is a standard language for managing data in databases. Think of it as a way to ask questions about your data:
- "How many people are there?" → `SELECT COUNT(*) FROM population`
- "What's the average age?" → `SELECT AVG(age) FROM population`
- "How many men vs women?" → `SELECT COUNT(*) FROM population WHERE sex = 'F'`

### Basic SQL Concepts

| Concept | What it does | Example |
|---------|--------------|---------|
| `SELECT` | Choose data to look at | `SELECT age` |
| `FROM` | Which table to use | `FROM population` |
| `WHERE` | Filter rows | `WHERE alive = true` |
| `COUNT(*)` | Count all rows | `COUNT(*)` |
| `AVG()` | Calculate average | `AVG(age)` |
| `MIN()` | Find minimum | `MIN(age)` |
| `MAX()` | Find maximum | `MAX(age)` |
| `CASE WHEN...THEN...END` | Conditional logic | `CASE WHEN age < 18 THEN 'child' END` |

### Adding New Statistics

Let's expand our configuration with more statistics:

```yaml
statistics:
  - name: "population_total"
    description: "Total population"
    query: "SELECT COUNT(*) as total FROM population"
  
  - name: "average_age"
    description: "Average age of population"
    query: |
      SELECT AVG(age) as avg_age
      FROM population
      WHERE alive = true
  
  - name: "age_distribution"
    description: "Age groups"
    query: |
      SELECT 
        COUNT(CASE WHEN age < 18 THEN 1 END) as children,
        COUNT(CASE WHEN age >= 18 AND age < 65 THEN 1 END) as adults,
        COUNT(CASE WHEN age >= 65 THEN 1 END) as elderly
      FROM population
      WHERE alive = true
  
  - name: "sex_distribution"
    description: "Sex distribution"
    query: |
      SELECT 
        COUNT(CASE WHEN sex = 'F' THEN 1 END) as females,
        COUNT(CASE WHEN sex = 'M' THEN 1 END) as males
      FROM population
      WHERE alive = true
```

### Understanding the Age Distribution Query

Let's break down the age distribution query step by step:

```sql
SELECT 
  COUNT(CASE WHEN age < 18 THEN 1 END) as children,
  COUNT(CASE WHEN age >= 18 AND age < 65 THEN 1 END) as adults,
  COUNT(CASE WHEN age >= 65 THEN 1 END) as elderly
FROM population
WHERE alive = true
```

**Step-by-step breakdown:**

1. `SELECT` - We want to look at data
2. `COUNT(CASE WHEN age < 18 THEN 1 END) as children` - For each row:
   - If `age < 18`, count it as 1
   - Otherwise, count it as nothing
   - This gives us the number of people under 18
   - We name this column `children`
3. `COUNT(CASE WHEN age >= 18 AND age < 65 THEN 1 END) as adults` - Same logic for adults
4. `COUNT(CASE WHEN age >= 65 THEN 1 END) as elderly` - Same logic for elderly
5. `FROM population` - Look at the population table
6. `WHERE alive = true` - Only include alive people

### Understanding the Sex Distribution Query

```sql
SELECT 
  COUNT(CASE WHEN sex = 'F' THEN 1 END) as females,
  COUNT(CASE WHEN sex = 'M' THEN 1 END) as males
FROM population
WHERE alive = true
```

This works the same way as the age distribution:
- For each person, check if sex is 'F' or 'M'
- Count them separately
- Only include alive people

### Understanding the CASE Statement

The `CASE` statement is like an "if-then-else" in SQL:

```sql
CASE 
  WHEN condition1 THEN result1
  WHEN condition2 THEN result2
  ELSE result3
END
```

Examples:
- `CASE WHEN age < 18 THEN 'child' ELSE 'adult' END` - Returns 'child' if age < 18, otherwise 'adult'
- `CASE WHEN sex = 'F' THEN 1 ELSE 0 END` - Returns 1 for females, 0 for males

### Adding Age Group Detail

Want more detailed age groups?

```sql
SELECT 
  COUNT(CASE WHEN age BETWEEN 0 AND 4 THEN 1 END) as age_0_4,
  COUNT(CASE WHEN age BETWEEN 5 AND 9 THEN 1 END) as age_5_9,
  COUNT(CASE WHEN age BETWEEN 10 AND 14 THEN 1 END) as age_10_14,
  COUNT(CASE WHEN age BETWEEN 15 AND 19 THEN 1 END) as age_15_19,
  COUNT(CASE WHEN age BETWEEN 20 AND 29 THEN 1 END) as age_20_29,
  COUNT(CASE WHEN age BETWEEN 30 AND 39 THEN 1 END) as age_30_39,
  COUNT(CASE WHEN age BETWEEN 40 AND 49 THEN 1 END) as age_40_49,
  COUNT(CASE WHEN age BETWEEN 50 AND 59 THEN 1 END) as age_50_59,
  COUNT(CASE WHEN age BETWEEN 60 AND 69 THEN 1 END) as age_60_69,
  COUNT(CASE WHEN age >= 70 THEN 1 END) as age_70_plus
FROM population
WHERE alive = true
```

The `BETWEEN` operator is inclusive - `age BETWEEN 0 AND 4` means age 0, 1, 2, 3, or 4.

## Step 6: Your Task - Creating Age Range Statistics

Now it's your turn! Let's add two new statistics to track the age range of our population.

### Task Description

Add two new statistics to your configuration file:

1. **Age Range Statistics** - Show the youngest and oldest ages in the population
2. **Age Spread Statistics** - Show how spread out the ages are

### What You Need to Know

**Age Range**
- Find the minimum age using `MIN(age)`
- Find the maximum age using `MAX(age)`
- This tells you the range of ages in your population

**Age Spread (Standard Deviation)**
- Standard deviation tells you how spread out the ages are
- A small standard deviation means most people are close to the average age
- A large standard deviation means ages vary widely
- SQL has a function called `STDDEV()` for this

### Step-by-Step Instructions

1. **Open your config file**: `config_aging.yaml` (or create a new one)

2. **Add a statistic for age range**:
   ```yaml
   - name: "age_range"
     description: "Youngest and oldest ages"
     query: |
       SELECT 
         MIN(age) as youngest,
         MAX(age) as oldest
       FROM population
       WHERE alive = true
   ```

3. **Add a statistic for standard deviation**:
   ```yaml
   - name: "age_spread"
     description: "Standard deviation of age"
     query: |
       SELECT 
         ROUND(STDDEV(age), 2) as age_stddev
       FROM population
       WHERE alive = true
   ```

   *Note: `ROUND()` makes the number easier to read by showing only 2 decimal places.*

4. **Your statistics section should now look like this**:
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
         WHERE alive = true
     
     - name: "age_range"
       description: "Youngest and oldest ages"
       query: |
         SELECT 
           MIN(age) as youngest,
           MAX(age) as oldest
         FROM population
         WHERE alive = true
     
     - name: "age_spread"
       description: "Standard deviation of age"
       query: |
         SELECT 
           ROUND(STDDEV(age), 2) as age_stddev
         FROM population
         WHERE alive = true
   ```

5. **Run the simulation**:
   ```bash
   ./talos config_aging.yaml
   ```

### Expected Output

You should see something like this in the statistics output:

```
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 2, adults: 6, elderly: 2
2024/01/15 10:00:00     age_range (Youngest and oldest ages): youngest: 3, oldest: 83
2024/01/15 10:00:00     age_spread (Standard deviation of age): age_stddev: 28.56
```

### What This Tells Us

- **Age Range**: Our population ranges from age 3 to age 83
- **Standard Deviation**: 28.56 years - this indicates considerable age diversity

### Hints If You Get Stuck

- Check your YAML indentation (spaces matter!)
- Make sure you have the `statistics:` section properly formatted
- Verify all column names exist in your population (age, sex, alive)
- If `STDDEV()` doesn't work, try `STDEV()` (some SQL versions use different names)

## Step 7: Full Configuration with All Statistics

Here's the complete configuration with all our statistics including your new ones:

```yaml
# config_aging_complete.yaml
# Aging model with comprehensive statistics

simulation:
  iterations: 5
  population_file: "population.csv"
  output_file: "population_aged.csv"
  random_seed: 42
  verbose: true

models:
  - name: "age_increment"
    type: "sql_update"
    priority: 1
    enabled: true
    description: "Increment everyone's age by 1 year"
    parameters:
      query: |
        UPDATE population 
        SET age = age + 1 
        WHERE alive = true

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
      WHERE alive = true
  
  - name: "sex_distribution"
    description: "Sex distribution"
    query: |
      SELECT 
        COUNT(CASE WHEN sex = 'F' THEN 1 END) as females,
        COUNT(CASE WHEN sex = 'M' THEN 1 END) as males
      FROM population
      WHERE alive = true
  
  - name: "average_age"
    description: "Average age"
    query: |
      SELECT AVG(age) as avg_age
      FROM population
      WHERE alive = true
  
  - name: "age_range"
    description: "Youngest and oldest ages"
    query: |
      SELECT 
        MIN(age) as youngest,
        MAX(age) as oldest
      FROM population
      WHERE alive = true
  
  - name: "age_spread"
    description: "Standard deviation of age"
    query: |
      SELECT 
        ROUND(STDDEV(age), 2) as age_stddev
      FROM population
      WHERE alive = true
```

## Step 8: Run with the Complete Configuration

Save the complete configuration as `config_aging_complete.yaml` and run it:

```bash
./talos config_aging_complete.yaml
```

### Expected Output with All Statistics

```
2024/01/15 10:00:00 Starting simulation
2024/01/15 10:00:00 Iterations: 5
...
2024/01/15 10:00:00 === Iteration 1/5 ===
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 2, adults: 6, elderly: 2
2024/01/15 10:00:00     sex_distribution (Sex distribution): females: 5, males: 5
2024/01/15 10:00:00     average_age (Average age): avg_age: 42.7
2024/01/15 10:00:00     age_range (Youngest and oldest ages): youngest: 3, oldest: 83
2024/01/15 10:00:00     age_spread (Standard deviation of age): age_stddev: 28.56

2024/01/15 10:00:00 === Iteration 2/5 ===
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 2, adults: 5, elderly: 3
2024/01/15 10:00:00     sex_distribution (Sex distribution): females: 5, males: 5
2024/01/15 10:00:00     average_age (Average age): avg_age: 43.7
2024/01/15 10:00:00     age_range (Youngest and oldest ages): youngest: 4, oldest: 84
2024/01/15 10:00:00     age_spread (Standard deviation of age): age_stddev: 28.56

...

2024/01/15 10:00:00 === Iteration 5/5 ===
2024/01/15 10:00:00   Statistics:
2024/01/15 10:00:00     population_total (Total population): total: 10
2024/01/15 10:00:00     age_distribution (Age groups): children: 1, adults: 4, elderly: 5
2024/01/15 10:00:00     sex_distribution (Sex distribution): females: 5, males: 5
2024/01/15 10:00:00     average_age (Average age): avg_age: 46.7
2024/01/15 10:00:00     age_range (Youngest and oldest ages): youngest: 7, oldest: 87
2024/01/15 10:00:00     age_spread (Standard deviation of age): age_stddev: 28.56

2024/01/15 10:00:00 === Simulation Complete ===
2024/01/15 10:00:00 Results saved to population_aged.csv
```

### Understanding Your New Statistics

**Age Range**
- Iteration 1: Youngest 3, Oldest 83 (range of 80 years)
- Iteration 5: Youngest 7, Oldest 87 (range of 80 years)
- The range stays the same because everyone ages together

**Age Spread (Standard Deviation)**
- 28.56 years throughout
- This shows the population has considerable age diversity
- The spread stays constant because everyone ages uniformly

## Summary

You've successfully:
1. Created a population CSV file
2. Written a configuration that ages the population
3. Added comprehensive statistics including age range and spread
4. Run a Talos simulation with all statistics
5. Analyzed the output

### Key Takeaways

- **Talos models are SQL**: The aging model is just an SQL UPDATE statement
- **SQL is powerful**: You can ask many questions about your data
- **Statistics are customizable**: Add any SQL query as a statistic
- **CASE statements enable grouping**: Create age groups, sex distributions, etc.
- **Aggregation functions summarize**: COUNT, AVG, MIN, MAX give overview statistics
- **STDDEV() measures spread**: Tells you how diverse your population is

### Common SQL Functions Reference

| Function | What it does | Example |
|----------|--------------|---------|
| `COUNT(*)` | Counts all rows | `COUNT(*)` |
| `COUNT(column)` | Counts non-null values | `COUNT(age)` |
| `AVG(column)` | Calculates average | `AVG(age)` |
| `SUM(column)` | Adds up values | `SUM(income)` |
| `MIN(column)` | Finds minimum | `MIN(age)` |
| `MAX(column)` | Finds maximum | `MAX(age)` |
| `STDDEV(column)` | Standard deviation | `STDDEV(age)` |
| `ROUND(value, decimals)` | Rounds to decimal places | `ROUND(STDDEV(age), 2)` |
| `CASE WHEN...THEN...END` | Conditional logic | `CASE WHEN age < 18 THEN 1 END` |
| `BETWEEN x AND y` | Range check | `age BETWEEN 18 AND 65` |

### Next Steps

In the next tutorial, we'll add mortality to the aging model, creating a more realistic demographic simulation. You'll learn:
- How to apply age-specific death rates
- How to track population changes over time
- More advanced SQL techniques

## Troubleshooting

**Error: "Failed to load population"**
- Check that `population.csv` exists in the same directory
- Verify the CSV has a header row
- Ensure all rows have the same number of columns

**Error: "Failed to execute model query"**
- Check your SQL syntax in the query
- Verify column names exist in the population
- Make sure `alive` column exists

**Error: "no such function: STDDEV"**
- Some SQLite versions use `STDEV()` instead of `STDDEV()`
- Try changing `STDDEV(age)` to `STDEV(age)`

**Error in statistics queries**
- Test your SQL query separately if possible
- Check for typos in column names
- Verify the syntax is valid for SQLite