package mcpimpl

import (
    "context"
    "fmt"
)

func HandleLeaveCalculator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    employeeID, _ :=getString(args, "employee_id")
    leaveDays, _ :=getInt(args, "leave_days")
    if leaveDays < 0 {
        return err("leave_days cannot be negative")
}

    msg := fmt.Sprintf("Employee %s requested %d leave days. Approved.", employeeID, leaveDays)
    return success(msg)
}

func HandlePayrollEstimator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    gross, _ :=getInt(args, "gross_salary")
    taxRate, _ :=getInt(args, "tax_rate")
    deductions, _ :=getInt(args, "deductions")
    if gross < 0 || taxRate < 0 || deductions < 0 {
        return err("all values must be non-negative")
}

    tax := gross * taxRate / 100
    net := gross - tax - deductions
    msg := fmt.Sprintf("Gross: %d, Tax: %d, Deductions: %d => Net Pay: %d", gross, tax, deductions, net)
    return success(msg)
}