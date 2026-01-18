package tasks

import (
	"testing"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestTransformTask_SimpleFieldExtraction(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	// Set up context with data
	ctx.Set("user", map[string]interface{}{
		"name": "Alice",
		"age":  30,
	})

	config := map[string]interface{}{
		"template":      "{{.name}}",
		"data_source":   "user",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
	assert.Equal(t, "Alice", result.Output)
}

func TestTransformTask_ArrayTransformation(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	users := []interface{}{
		map[string]interface{}{"name": "Alice"},
		map[string]interface{}{"name": "Bob"},
		map[string]interface{}{"name": "Charlie"},
	}
	ctx.Set("users_list", map[string]interface{}{"users": users})

	config := map[string]interface{}{
		"template":      `[{{range $i, $u := .users}}{{if $i}},{{end}}"{{$u.name}}"{{end}}]`,
		"data_source":   "users_list",
		"output_format": "json",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)

	// Parse output as array
	outputArray, ok := result.Output.([]interface{})
	assert.True(t, ok)
	assert.Len(t, outputArray, 3)
	assert.Equal(t, "Alice", outputArray[0])
	assert.Equal(t, "Bob", outputArray[1])
	assert.Equal(t, "Charlie", outputArray[2])
}

func TestTransformTask_DataReshaping(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("api_response", map[string]interface{}{
		"data": map[string]interface{}{
			"firstName": "Alice",
			"lastName":  "Smith",
			"age":       30,
		},
	})

	config := map[string]interface{}{
		"template":      `{"fullName": "{{.data.firstName}} {{.data.lastName}}", "age": {{.data.age}}}`,
		"data_source":   "api_response",
		"output_format": "json",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)

	output, ok := result.Output.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Alice Smith", output["fullName"])
	assert.Equal(t, float64(30), output["age"]) // JSON numbers are float64
}

func TestTransformTask_CustomFunction_ToUpper(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "alice",
	})

	config := map[string]interface{}{
		"template":      `{{.name | toUpper}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
	assert.Equal(t, "ALICE", result.Output)
}

func TestTransformTask_CustomFunction_ToLower(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "ALICE",
	})

	config := map[string]interface{}{
		"template":      `{{.name | toLower}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "alice", result.Output)
}

func TestTransformTask_CustomFunction_Trim(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"text": "  hello world  ",
	})

	config := map[string]interface{}{
		"template":      `{{.text | trim}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "hello world", result.Output)
}

func TestTransformTask_CustomFunction_Join(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	items := []interface{}{"apple", "banana", "cherry"}
	ctx.Set("items", map[string]interface{}{"list": items})

	config := map[string]interface{}{
		"template":      `{{.list | join ", "}}`,
		"data_source":   "items",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
	assert.Equal(t, "apple, banana, cherry", result.Output)
}

func TestTransformTask_CustomFunction_ToJSON(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  30,
		},
	})

	config := map[string]interface{}{
		"template":      `{{.user | toJSON}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Contains(t, result.Output, `"name":"Alice"`)
	assert.Contains(t, result.Output, `"age":30`)
}

func TestTransformTask_CustomFunction_Default(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "",
	})

	config := map[string]interface{}{
		"template":      `{{.name | default "Unknown"}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "Unknown", result.Output)
}

func TestTransformTask_CustomFunction_DefaultWithValue(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "Alice",
	})

	config := map[string]interface{}{
		"template":      `{{.name | default "Unknown"}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "Alice", result.Output)
}

func TestTransformTask_AllContextAccess(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	// Set multiple keys in context
	ctx.Set("user", map[string]interface{}{"name": "Alice"})
	ctx.Set("order", map[string]interface{}{"total": 100})

	config := map[string]interface{}{
		"template": `{"customer": "{{.user.name}}", "amount": {{.order.total}}}`,
		// No data_source - access entire context
		"output_format": "json",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)

	output, ok := result.Output.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Alice", output["customer"])
	assert.Equal(t, float64(100), output["amount"])
}

func TestTransformTask_InvalidTemplate(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		"template": `{{.invalid syntax`,
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "failed to parse template")
}

func TestTransformTask_MissingTemplate(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		// template missing
		"data_source": "something",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "template")
}

func TestTransformTask_MissingDataSource(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	// Don't set "nonexistent" in context

	config := map[string]interface{}{
		"template":      "{{.name}}",
		"data_source":   "nonexistent",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	// Should succeed but with empty data (graceful handling)
	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
}

func TestTransformTask_TemplateExecutionError(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"value": 123,
	})

	// This template will fail because we're trying to call a string method on a number
	config := map[string]interface{}{
		"template":      `{{.value | toUpper}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "failed to execute template")
}

func TestTransformTask_JSONOutputFormat(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "Alice",
		"age":  30,
	})

	config := map[string]interface{}{
		"template":      `{"name": "{{.name}}", "age": {{.age}}}`,
		"data_source":   "data",
		"output_format": "json",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	output, ok := result.Output.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Alice", output["name"])
	assert.Equal(t, float64(30), output["age"])
}

func TestTransformTask_StringOutputFormat(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "Alice",
		"age":  30,
	})

	config := map[string]interface{}{
		"template":      `Name: {{.name}}, Age: {{.age}}`,
		"data_source":   "data",
		"output_format": "string",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "Name: Alice, Age: 30", result.Output)
}

func TestTransformTask_InvalidJSON_ReturnsAsString(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("data", map[string]interface{}{
		"name": "Alice",
	})

	config := map[string]interface{}{
		"template":      `This is not JSON: {{.name}}`,
		"data_source":   "data",
		"output_format": "json",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	// Should return as string since it's not valid JSON
	assert.Equal(t, "This is not JSON: Alice", result.Output)
}

func TestTransformTask_ComplexNesting(t *testing.T) {
	task := NewTransformTask()
	ctx := engine.NewExecutionContext()

	ctx.Set("response", map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
				map[string]interface{}{"id": 2, "name": "Bob"},
			},
		},
	})

	config := map[string]interface{}{
		"template":      `[{{range $i, $u := .data.users}}{{if $i}},{{end}}{"id":{{$u.id}},"name":"{{$u.name}}"}{{end}}]`,
		"data_source":   "response",
		"output_format": "json",
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	outputArray, ok := result.Output.([]interface{})
	assert.True(t, ok)
	assert.Len(t, outputArray, 2)

	user1 := outputArray[0].(map[string]interface{})
	assert.Equal(t, float64(1), user1["id"])
	assert.Equal(t, "Alice", user1["name"])
}

func TestTransformTask_RegisterTransformTask(t *testing.T) {
	registry := engine.NewRegistry()

	RegisterTransformTask(registry)

	executor, err := registry.Get("transform")
	assert.NoError(t, err)
	assert.NotNil(t, executor)
	assert.IsType(t, &TransformTask{}, executor)
}

func TestTransformTask_NewTransformTask(t *testing.T) {
	task := NewTransformTask()

	assert.NotNil(t, task)
	assert.NotNil(t, task.funcMap)

	// Verify custom functions are registered
	_, hasToUpper := task.funcMap["toUpper"]
	assert.True(t, hasToUpper)

	_, hasToLower := task.funcMap["toLower"]
	assert.True(t, hasToLower)

	_, hasTrim := task.funcMap["trim"]
	assert.True(t, hasTrim)

	_, hasJoin := task.funcMap["join"]
	assert.True(t, hasJoin)

	_, hasToJSON := task.funcMap["toJSON"]
	assert.True(t, hasToJSON)

	_, hasDefault := task.funcMap["default"]
	assert.True(t, hasDefault)
}
