# Talos Tutorial 1: Building an Aging Model

## Overview

In this tutorial, you'll learn how to create a simple aging model with Talos. We'll start with a CSV population file, write a configuration that ages everyone by one year, run the simulation, and analyze the output.

## Prerequisites

- Talos binary downloaded and in your PATH (or in the current directory)
- Basic understanding of CSV files
- A text editor (VS Code, Sublime, Notepad++, etc.)

## What You'll Learn

By the end of this tutorial, you'll be able to:
- Create a population CSV file
- Write a YAML configuration file
- Run a Talos simulation
- Add custom statistics using simple SQL queries

**Important Note:** You don't need to understand SQL databases to use Talos! We're only using a small subset of SQL for **data retrieval** (asking questions about your data) and **data manipulation** (making changes to your data). Think of it as a way to:
- **Ask**: "How many people are under 18?" → `COUNT(CASE WHEN age < 18 THEN 1 END)`
- **Change**: "Make everyone one year older" → `UPDATE population SET age = age + 1`

No database administration, no complex queries, no database design - just simple statements that work like plain English!

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

## Step 2: Understanding YAML Configuration Files

Before we create our configuration file, let's understand the YAML format. YAML (YAML Ain't Markup Language) is a human-readable data format that Talos uses for configuration.

### Important YAML Rules

**1. Indentation Matters (Spaces Only, No Tabs!)**

YAML uses indentation to show structure. Always use spaces, never tabs. The number of spaces doesn't matter as long as it's consistent, but 2 spaces is the standard convention.

```yaml
# ✅ CORRECT - using spaces
simulation:
  iterations: 5
  population_file: "population.csv"

# ❌ WRONG - using tabs (invisible but will cause errors)
simulation:
	iterations: 5
	population_file: "population.csv"

# ❌ WRONG - inconsistent indentation
simulation:
    iterations: 5
  population_file: "population.csv"
```

**2. Case Sensitivity**

YAML is case-sensitive. `simulation` is not the same as `Simulation`.

```yaml
# ✅ CORRECT
simulation:
  iterations: 5

# ❌ WRONG - uppercase S
Simulation:
  iterations: 5
```

**3. Column Names Must Match Your CSV**

This is the most important rule for Talos! The column names you use in your SQL queries must exactly match the column names in your CSV file.

```csv
# CSV column names
person_id,age,sex,area,alive
```

```sql
-- ✅ CORRECT - matches CSV
UPDATE population 
SET age = age + 1 
WHERE alive = true

-- ❌ WRONG - 'Age' doesn't match 'age' in CSV
UPDATE population 
SET Age = Age + 1 
WHERE Alive = true
```

**4. Key-Value Pairs**

YAML uses `key: value` format with a colon and a space:

```yaml
# ✅ CORRECT
name: "age_increment"
priority: 1

# ❌ WRONG - missing space after colon
name:"age_increment"
priority:1
```

**5. Multi-line Strings with the Pipe (`|`)**

For SQL queries and long text, use the pipe (`|`) character for multi-line strings:

```yaml
# ✅ CORRECT - preserves line breaks
query: |
  UPDATE population 
  SET age = age + 1 
  WHERE alive = true

# ❌ WRONG - single line is hard to read
query: "UPDATE population SET age = age + 1 WHERE alive = true"
```

**Why do we need the pipe (`|`)?**

The pipe tells YAML that the following text is a **multi-line string**. Without it, YAML would treat each line as a separate value or would combine everything into a single line.

**Let's see what happens without the pipe:**

```yaml
# ❌ WRONG - YAML will fail to parse this
query: UPDATE population 
SET age = age + 1 
WHERE alive = true
```

YAML would see this as three separate lines and throw an error because it expects a single value after the colon.

**With the pipe (`|`):**

```yaml
# ✅ CORRECT - YAML preserves the multi-line format
query: |
  UPDATE population 
  SET age = age + 1 
  WHERE alive = true
```

The pipe tells YAML: "Everything that follows, indented, is a single string value." This allows us to write clean, readable SQL queries that span multiple lines.

**What about other multi-line options?**

| Option | What it does | When to use |
|--------|--------------|-------------|
| `|` (pipe) | Preserves line breaks | For SQL queries where formatting matters |
| `>` (greater than) | Folds lines into one | For long text descriptions |
| `|-` (pipe with dash) | Same as `|` but strips trailing newline | For cleaner output |

**Indentation matters with the pipe:**

The content after the pipe must be indented (usually 2 or 4 spaces):

```yaml
# ✅ CORRECT - content is indented
query: |
  UPDATE population 
  SET age = age + 1 
  WHERE alive = true

# ❌ WRONG - content is not indented
query: |
UPDATE population 
SET age = age + 1 
WHERE alive = true
```

**Quick Summary:**

| Feature | Without \| | With \| |
|---------|------------|---------|
| Lines allowed | Single line only | Multiple lines allowed |
| Example | `query: "UPDATE population SET age = age + 1 WHERE alive = true"` | `query: \|`<br>`  UPDATE population`<br>`  SET age = age + 1`<br>`  WHERE alive = true` |
| Readability | Hard to read for complex queries | Easy to read and maintain |

**Remember:**
- Always use the pipe (`|`) for SQL queries that span multiple lines
- Content after the pipe must be indented
- This is a YAML rule, not a Talos rule!

**6. Lists (Arrays)**

Use a dash (`-`) for list items:

```yaml
# ✅ CORRECT
models:
  - name: "age_increment"
    priority: 1
  - name: "mortality"
    priority: 2

# ❌ WRONG - missing dash
models:
  name: "age_increment"
  priority: 1
  name: "mortality"
  priority: 2
```

**7. Common YAML Mistakes**

| Mistake | Example | Fix |
|---------|---------|-----|
| Missing colon | `name "age_increment"` | `name: "age_increment"` |
| Missing space after colon | `name:"age_increment"` | `name: "age_increment"` |
| Using tabs | `\titerations: 5` | Use spaces instead |
| Inconsistent indentation | `iterations:\n 5` | Use consistent spaces |
| Case mismatch | `Simulation:` | `simulation:` |
| Column name mismatch | `Age` in query, `age` in CSV | Match exactly |
| Missing pipe for multi-line | `query: UPDATE...` | `query: |` then new line |

### Basic YAML Structure for Talos

Here's the basic structure of a Talos configuration file:

```yaml
simulation:                   # Level 1
  iterations: 5              # Level 2
  population_file: "file"    # Level 2

models:                       # Level 1
  - name: "model1"           # Level 2 (list item)
    type: "sql_update"       # Level 3
    priority: 1              # Level 3
    parameters:              # Level 3
      query: |               # Level 4 (pipe for multi-line)
        UPDATE population    # Level 4 continues
        SET age = age + 1    # Level 4 continues
```

## Step 3: Create the Configuration File

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

**Important: The column names in the SQL query must match your CSV exactly!**

In our CSV we have:
- `age` (lowercase)
- `alive` (lowercase)

So our SQL uses:
- `age` (lowercase)
- `alive` (lowercase)

If your CSV had `Age` and `Alive` (capitalized), you'd need to use those exact names.

**Statistics Section:**
- We define two statistics to track
- `population_total` - Counts everyone
- `age_distribution` - Groups population into children, adults, and elderly

### Common YAML Errors to Avoid

1. **Mismatched column names:**
   ```yaml
   # If your CSV has 'age' but you write:
   UPDATE population SET Age = Age + 1  # ❌ WRONG
   UPDATE population SET age = age + 1  # ✅ CORRECT
   ```

2. **Incorrect indentation:**
   ```yaml
   models:
   - name: "age_increment"
   type: "sql_update"  # ❌ WRONG - should be indented
   ```

3. **Missing or extra spaces:**
   ```yaml
   name:"age_increment"  # ❌ WRONG - missing space
   name: "age_increment" # ✅ CORRECT
   ```

4. **Mixing tabs and spaces:**
   - Always use spaces. Most text editors can show invisible characters.

## Step 4: Run the Simulation

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

## Step 5: Examine the Output

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

## Step 6: Adding More Statistics (Understanding SQL)

Now let's add more statistics to better understand our population. **This is where we'll learn the basics of SQL**, which is the language Talos uses for models and statistics.

### Don't Worry - You Don't Need to Be a Database Expert!

SQL (Structured Query Language) might sound intimidating, but **you only need to learn a few simple patterns** to use Talos effectively. Think of it like learning a few phrases in a new language - you don't need to be fluent!

**The SQL we use in Talos is just for two things:**

1. **Asking questions** (retrieving data):
   - "How many people are there?" → `SELECT COUNT(*) FROM population`
   - "What's the average age?" → `SELECT AVG(age) FROM population`
   - "How many children are there?" → `SELECT COUNT(*) FROM population WHERE age < 18`

2. **Making changes** (manipulating data):
   - "Make everyone one year older" → `UPDATE population SET age = age + 1`
   - "Only update alive people" → `UPDATE population SET age = age + 1 WHERE alive = true`

**What you DON'T need to know:**
- ❌ How to create or manage databases
- ❌ How to design database schemas
- ❌ How to optimize queries for performance
- ❌ How to use joins or complex subqueries
- ❌ How to administer a database system

**What you DO need to know:**
- ✅ How to ask questions about your data (SELECT)
- ✅ How to make changes to your data (UPDATE)
- ✅ How to filter data (WHERE)
- ✅ How to count and summarize (COUNT, AVG, MIN, MAX)
- ✅ How to group data conditionally (CASE WHEN...THEN)

And that's it! The examples in this tutorial cover everything you'll need for most demographic simulations.

### What is SQL?

SQL (Structured Query Language) is a standard language for managing data in databases. Think of it as a way to ask questions about your data:
- "How many people are there?" → `SELECT COUNT(*) FROM population`
- "What's the average age?" → `SELECT AVG(age) FROM population`
- "How many men vs women?" → `SELECT COUNT(*) FROM population WHERE sex = 'F'`

**In Talos, SQL is just a tool - not the focus.** You're using it to tell the simulation engine what to do. The engine handles all the complex database stuff behind the scenes. You just write simple statements that read like English.

Think of it like using a calculator:
- You don't need to understand how the calculator works internally
- You just need to know which buttons to press
- The calculator does the heavy lifting

Same with SQL in Talos - you just need to know the few patterns we show you!

### Basic SQL Concepts

Remember: **These are the ONLY concepts you need to know!** Everything else is handled by Talos.

| Concept | What it does | Example | Plain English |
|---------|--------------|---------|---------------|
| `SELECT` | Choose data to look at | `SELECT age` | "Show me the ages" |
| `FROM` | Which table to use | `FROM population` | "From the population data" |
| `WHERE` | Filter rows | `WHERE alive = true` | "Only where alive is true" |
| `COUNT(*)` | Count all rows | `COUNT(*)` | "Count everyone" |
| `AVG()` | Calculate average | `AVG(age)` | "Average age" |
| `MIN()` | Find minimum | `MIN(age)` | "Youngest person" |
| `MAX()` | Find maximum | `MAX(age)` | "Oldest person" |
| `CASE WHEN...THEN...END` | Conditional logic | `CASE WHEN age < 18 THEN 1 END` | "If age is under 18, count it" |

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

### Understanding Female Children Example

If you want to count **female children** specifically, you add `AND sex = 'F'`:

```sql
COUNT(CASE WHEN age < 18 AND sex = 'F' THEN 1 END) as female_children
```

Similarly for other groups:

```sql
SELECT 
  -- Female children
  COUNT(CASE WHEN age < 18 AND sex = 'F' THEN 1 END) as female_children,
  -- Male children
  COUNT(CASE WHEN age < 18 AND sex = 'M' THEN 1 END) as male_children,
  -- Female adults
  COUNT(CASE WHEN age >= 18 AND age < 65 AND sex = 'F' THEN 1 END) as female_adults,
  -- Male adults
  COUNT(CASE WHEN age >= 18 AND age < 65 AND sex = 'M' THEN 1 END) as male_adults,
  -- Female elderly
  COUNT(CASE WHEN age >= 65 AND sex = 'F' THEN 1 END) as female_elderly,
  -- Male elderly
  COUNT(CASE WHEN age >= 65 AND sex = 'M' THEN 1 END) as male_elderly
FROM population
WHERE alive = true
```

**Key point:** Column names are **case-sensitive** and must match your CSV. If your CSV has `sex` (lowercase), use `sex`. If it has `Sex` (uppercase S), use `Sex`.

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

## Step 7: Your Task - Creating Age Range Statistics

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

## Step 8: Full Configuration with All Statistics

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

## Step 9: Run with the Complete Configuration

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
2. Learned about YAML format and its rules (including the pipe `|` for multi-line strings)
3. Written a configuration that ages the population
4. Added comprehensive statistics including age range and spread
5. Run a Talos simulation with all statistics
6. Analyzed the output

### Key Takeaways

- **You don't need to be a database expert**: Talos uses a small, friendly subset of SQL
- **YAML is whitespace-sensitive**: Indentation matters and always use spaces, never tabs
- **The pipe (`|`) is for multi-line strings**: Essential for writing readable SQL queries
- **Column names must match exactly**: What you put in your CSV must match what you write in SQL queries
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
- Make sure column names match your CSV exactly (case-sensitive!)
- Make sure `alive` column exists

**Error: "no such function: STDDEV"**
- Some SQLite versions use `STDEV()` instead of `STDDEV()`
- Try changing `STDDEV(age)` to `STDEV(age)`

**Error in statistics queries**
- Test your SQL query separately if possible
- Check for typos in column names
- Verify the syntax is valid for SQLite

**YAML parsing errors**
- Check your indentation (use spaces, not tabs)
- Make sure there's a space after each colon
- Verify all quotes are balanced
- Use a YAML validator online to check your syntax
- **Remember the pipe (`|`) for multi-line SQL queries!**
