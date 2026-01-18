package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockExecutor_Success(t *testing.T) {
	executor := &MockExecutor{
		ShouldFail: false,
		Output:     map[string]interface{}{"data": "test"},
	}

	ctx := NewExecutionContext()
	config := map[string]interface{}{}

	result := executor.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.NotNil(t, result.Output)
	assert.Empty(t, result.Error)
	assert.Equal(t, map[string]interface{}{"data": "test"}, result.Output)
}

func TestMockExecutor_SuccessWithDefaultOutput(t *testing.T) {
	executor := &MockExecutor{
		ShouldFail: false,
		Output:     nil, // Should use default mock result
	}

	ctx := NewExecutionContext()
	result := executor.Execute(ctx, map[string]interface{}{})

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, map[string]interface{}{"mock": "result"}, result.Output)
	assert.Empty(t, result.Error)
}

func TestMockExecutor_Failure(t *testing.T) {
	executor := &MockExecutor{
		ShouldFail: true,
		ErrorMsg:   "mock error",
	}

	ctx := NewExecutionContext()
	result := executor.Execute(ctx, map[string]interface{}{})

	assert.Equal(t, "failed", result.Status)
	assert.Nil(t, result.Output)
	assert.Equal(t, "mock error", result.Error)
}

func TestMockExecutor_ImplementsInterface(t *testing.T) {
	var _ TaskExecutor = (*MockExecutor)(nil)
}

func TestMockExecutor_WithConfig(t *testing.T) {
	executor := &MockExecutor{
		ShouldFail: false,
		Output:     map[string]interface{}{"processed": "data"},
	}

	ctx := NewExecutionContext()
	config := map[string]interface{}{
		"param1": "value1",
		"param2": 123,
	}

	result := executor.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.NotNil(t, result.Output)
}
