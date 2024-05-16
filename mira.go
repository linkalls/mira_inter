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

// Error method for CustomError
func (e CustomError) Error() string {
    // エラーメッセージの基本フレームワークを短縮
    baseMsg := "Error at line %d: Unexpected error"
    
    // エラーメッセージの詳細部分を追加
    detailMsg := ""
    if strings.Contains(e.Message, "Invalid expression:") {
        detailMsg = "Invalid expression detected."
    } else if strings.Contains(e.Message, "Function not found:") {
        detailMsg = "Undefined function called."
    } else {
        detailMsg = "Unexpected error."
    }
    
    // 最終的なエラーメッセージを組み立てる
    finalMsg := fmt.Sprintf(baseMsg+": %s", detailMsg)
    
    return finalMsg
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

    // Try to perform an arithmetic operation
    if strings.ContainsAny(expr, "+-*/") {
        parts := strings.Fields(expr)
        if len(parts) != 3 {
            return nil, CustomError{Message: "Invalid expression: " + expr, Line: line}
        }

        a, errA := strconv.ParseFloat(parts[0], 64)
        b, errB := strconv.ParseFloat(parts[2], 64)

        if parts[1] == "+" && (errA != nil || errB != nil) {
            // If the operation is addition and either of the operands is not a number,
            // treat them as strings and concatenate them
            return parts[0] + parts[2], nil
        }

        if errA != nil {
            return nil, CustomError{Message: "Invalid number: " + parts[0], Line: line}
        }

        if errB != nil {
            return nil, CustomError{Message: "Invalid number: " + parts[2], Line: line}
        }

        result, err := performArithmeticOperation(parts[1], a, b)
        if err != nil {
            return nil, CustomError{Message: err.Error(), Line: line}
        }

        return result, nil
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

        ExecuteStatement(line, i+1, false) // Pass false for the interactive parameter
    }
}

// ExecuteStatement 関数
func ExecuteStatement(statement string, line int, interactive bool) {
    defer func() {
        if r := recover(); r!= nil {
            if interactive {
                fmt.Println("=> undefined")
            } else {
                panic(r)
            }
        }
    }()

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
        if interactive {
            fmt.Println("=>", varValue) // Add this line
        }
    } else if strings.HasPrefix(statement, "fn ") {
        defineFunction(statement, line)
    } else {
        result, err := evaluateExpression(statement, line)
        if err != nil {
            panic(CustomError{Message: err.Error(), Line: line})
        }
        if interactive {
            fmt.Println("=>", result) // Add this line
        }
    }
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
        ExecuteStatement(input, 0, true)
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
        fmt.Println("Mira version 0.01")
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


// Function to perform arithmetic operations
func performArithmeticOperation(op string, a, b float64) (float64, error) {
    switch op {
    case "+":
        return a + b, nil
    case "-":
        return a - b, nil
    case "*":
        return a * b, nil
    case "/":
        if b == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("unknown operator: %s", op)
    }
}

// Function to evaluate an arithmetic expression
func evaluateArithmeticExpression(expr string) (float64, error) {
    parts := strings.Fields(expr)
    if len(parts) != 3 {
        return 0, fmt.Errorf("invalid expression: %s", expr)
    }

    a, err := strconv.ParseFloat(parts[0], 64)
    if err != nil {
        return 0, fmt.Errorf("invalid number: %s", parts[0])
    }

    b, err := strconv.ParseFloat(parts[2], 64)
    if err != nil {
        return 0, fmt.Errorf("invalid number: %s", parts[2])
    }

    return performArithmeticOperation(parts[1], a, b)
}



