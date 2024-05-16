package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)

// CustomError type
type CustomError struct {
    Message string
    Line    int
}

func (e CustomError) Error() string {
    return fmt.Sprintf("Error at line %d: %s", e.Line, e.Message)
}

// Variable to store variables and functions
var variables = make(map[string]interface{})
var functions = make(map[string]func([]interface{}) interface{})

// Function to evaluate an expression
func evaluateExpression(expr string, line int) (interface{}, error) {
    // Try to get the value of a variable or function call
    if strings.Contains(expr, "(") {
        parts := strings.SplitN(expr, "(", 2)
        functionName := strings.TrimSpace(parts[0])
        argsExpr := strings.TrimSpace(strings.TrimSuffix(parts[1], ")"))
        args, err := evaluateArguments(argsExpr, line)
        if err != nil {
            return nil, err
        }
        if function, ok := functions[functionName]; ok {
            return function(args), nil
        }
        return nil, CustomError{Message: "Function not found: " + functionName, Line: line}
    }

    // Check if the expression is a variable name
    if value, ok := variables[expr]; ok {
        return value, nil
    }

    // Try to convert the expression to an integer
    if intValue, err := strconv.Atoi(expr); err == nil {
        return intValue, nil
    }

    // Try to convert the expression to a float
    if floatValue, err := strconv.ParseFloat(expr, 64); err == nil {
        return floatValue, nil
    }
    // Check if the expression is a quoted string
    if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
        strValue := strings.Trim(expr, "\"")
        // Replace #{variable} with actual variable values
        re := regexp.MustCompile(`#{([^}]+)}`)
        strValue = re.ReplaceAllStringFunc(strValue, func(match string) string {
            varName := match[2 : len(match)-1]
            if value, ok := variables[varName]; ok {
                return fmt.Sprintf("%v", value)
            }
            return match
        })
        // Unquote the string to interpret escape sequences
        unquotedStrValue, err := strconv.Unquote(`"` + strValue + `"`)
        if err != nil {
            return nil, CustomError{Message: "Invalid string: " + expr, Line: line}
        }
        return unquotedStrValue, nil
    }
    
    return nil, CustomError{Message: "Invalid expression: " + expr, Line: line}
} // <- Add this closing bracket

    

// Function to evaluate a list of arguments
func evaluateArguments(argsExpr string, line int) ([]interface{}, error) {
    args := []interface{}{}
    if argsExpr != "" {
        argsList := strings.Split(argsExpr, ",")
        for _, argExpr := range argsList {
            argExpr = strings.TrimSpace(argExpr)
            value, err := evaluateExpression(argExpr, line)
            if err != nil {
                return nil, err
            }
            args = append(args, value)
        }
    }
    return args, nil
}

// Function to define a user-defined function
func defineFunction(statement string, line int) {
    parts := strings.SplitN(statement, "=", 2)
    functionName := strings.TrimSpace(parts[0])
    functionBody := strings.TrimSpace(strings.TrimPrefix(parts[1], "fn("))
    functions[functionName] = func(args []interface{}) interface{} {
        for i, arg := range args {
            variables[functionBody[i:i+1]] = arg
        }
        result, _ := evaluateExpression(functionBody, line)
        return result
    }
}

// ExecuteScript function
func ExecuteScript(script string) {
    lines := strings.Split(script, "\n")

    for i, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        ExecuteStatement(line, i+1)
    }
}

// ExecuteStatement function
func ExecuteStatement(statement string, line int) {
    if strings.HasPrefix(statement, "puts ") {
        message, err := evaluateExpression(strings.TrimPrefix(statement, "puts "), line)
        if err != nil {
            panic(CustomError{Message: err.Error(), Line: line})
        }
        fmt.Println(message)
    } else if strings.HasPrefix(statement, "print ") {
        message, err := evaluateExpression(strings.TrimPrefix(statement, "print "), line)
        if err != nil {
            panic(CustomError{Message: err.Error(), Line: line})
        }
        fmt.Print(message)
    } else if strings.HasPrefix(statement, "p ") {
        message, err := evaluateExpression(strings.TrimPrefix(statement, "p "), line)
        if err != nil {
            panic(CustomError{Message: err.Error(), Line: line})
        }
        fmt.Println(fmt.Sprintf("%v", message))
    } else if strings.Contains(statement, "=") {
        parts := strings.SplitN(statement, "=", 2)
        varName := strings.TrimSpace(parts[0])
        varValue, err := evaluateExpression(strings.TrimSpace(parts[1]), line)
        if err != nil {
            panic(CustomError{Message: err.Error(), Line: line})
        }
        variables[varName] = varValue
    } else if strings.HasPrefix(statement, "fn ") {
        defineFunction(statement, line)
    } else {
        _, err := evaluateExpression(statement, line)
        if err != nil {
            panic(CustomError{Message: err.Error(), Line: line})
        }
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: mira <script.mr> or mira -v")
        os.Exit(1)
    }

    if os.Args[1] == "-cmd" {
        startInteractiveMode()
        os.Exit(0)
    }

    if os.Args[1] == "-v" || os.Args[1] == "--version" {
        fmt.Println("Mira version 0.4")
        os.Exit(0)
    }

    scriptPath := os.Args[1]
    file, err := os.Open(scriptPath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        os.Exit(1)
    }
    defer file.Close()

    // Read script content
    scanner := bufio.NewScanner(file)
    var scriptContent string
    for scanner.Scan() {
        scriptContent += scanner.Text() + "\n"
    }
    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
        os.Exit(1)
    }

    
    // Execute script logic
    ExecuteScript(scriptContent)
}

func startInteractiveMode() {
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("> ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)
        if input == "exit" {
            break
        }
        ExecuteStatement(input, 0)
    }
}

