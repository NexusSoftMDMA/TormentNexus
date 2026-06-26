package tools

import "context"

func HandlePhoneLookup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phone, _ :=getString(args, "phone")
	if phone == "" {
		return err("phone number is required")
}

	return success("Phone number: " + phone)
}

func HandlePhoneValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	phone, _ :=getString(args, "phone")
	if phone == "" {
		return err("phone number is required")
}

	return success("Phone number is valid")
}