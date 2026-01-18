package tasks

import (
	"testing"

	"github.com/davioliveira/rest_api_automation_hub_go/internal/engine"
	"github.com/stretchr/testify/assert"
)

const sampleHTML = `
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<h1>Welcome</h1>
	<div class="container">
		<p class="description">This is a test page</p>
		<ul class="items">
			<li>Item 1</li>
			<li>Item 2</li>
			<li>Item 3</li>
		</ul>
		<a href="/page1" class="link">Link 1</a>
		<a href="/page2" class="link">Link 2</a>
		<img src="/image1.jpg" alt="Image 1">
		<img src="/image2.jpg" alt="Image 2">
	</div>
</body>
</html>
`

func TestHTMLParserTask_SingleElementExtraction(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)
	assert.NotNil(t, result.Output)

	results := result.Output.([]map[string]any)
	assert.Len(t, results, 1)
	assert.Equal(t, "Welcome", results[0]["title"])
}

func TestHTMLParserTask_MultipleElementExtraction(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "items",
				"selector": ".items li",
				"multiple": true,
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)
	assert.Empty(t, result.Error)

	results := result.Output.([]map[string]any)
	assert.Len(t, results, 1)

	items := results[0]["items"].([]string)
	assert.Len(t, items, 3)
	assert.Equal(t, "Item 1", items[0])
	assert.Equal(t, "Item 2", items[1])
	assert.Equal(t, "Item 3", items[2])
}

func TestHTMLParserTask_AttributeExtraction(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":      "first_link",
				"selector":  "a.link",
				"attribute": "href",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	assert.Equal(t, "/page1", results[0]["first_link"])
}

func TestHTMLParserTask_MultipleAttributeExtraction(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":      "links",
				"selector":  "a.link",
				"attribute": "href",
				"multiple":  true,
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	links := results[0]["links"].([]string)
	assert.Len(t, links, 2)
	assert.Equal(t, "/page1", links[0])
	assert.Equal(t, "/page2", links[1])
}

func TestHTMLParserTask_ImageSourceExtraction(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":      "images",
				"selector":  "img",
				"attribute": "src",
				"multiple":  true,
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	images := results[0]["images"].([]string)
	assert.Len(t, images, 2)
	assert.Equal(t, "/image1.jpg", images[0])
	assert.Equal(t, "/image2.jpg", images[1])
}

func TestHTMLParserTask_MultipleSelectors(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
			map[string]interface{}{
				"name":     "description",
				"selector": ".description",
			},
			map[string]interface{}{
				"name":     "items",
				"selector": ".items li",
				"multiple": true,
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	assert.Len(t, results, 1)

	assert.Equal(t, "Welcome", results[0]["title"])
	assert.Equal(t, "This is a test page", results[0]["description"])

	items := results[0]["items"].([]string)
	assert.Len(t, items, 3)
}

func TestHTMLParserTask_NestedSelectors(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "container_text",
				"selector": ".container .description",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	assert.Equal(t, "This is a test page", results[0]["container_text"])
}

func TestHTMLParserTask_EmptyResults(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "nonexistent",
				"selector": ".does-not-exist",
			},
		},
	}

	result := task.Execute(ctx, config)

	// Should succeed but with empty value
	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	assert.Equal(t, "", results[0]["nonexistent"])
}

func TestHTMLParserTask_MissingHTMLSource(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	// Don't set html_content in context

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "not found in context")
}

func TestHTMLParserTask_InvalidHTMLSourceConfig(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	config := map[string]interface{}{
		// html_source missing
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "html_source")
}

func TestHTMLParserTask_MissingSelectors(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		// selectors missing
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "selectors")
}

func TestHTMLParserTask_InvalidSelectorConfig_MissingName(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				// name missing
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "missing 'name'")
}

func TestHTMLParserTask_InvalidSelectorConfig_MissingSelector(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", sampleHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name": "title",
				// selector missing
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "missing 'selector'")
}

func TestHTMLParserTask_InvalidHTML(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	// Goquery is very forgiving, but let's test with malformed HTML
	ctx.Set("html_content", "<<<>>>invalid")

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	// Goquery will parse even malformed HTML successfully
	assert.Equal(t, "success", result.Status)
}

func TestHTMLParserTask_HTMLFromHTTPResponse(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	// Simulate HTTP response structure
	httpResponse := map[string]interface{}{
		"status_code": 200,
		"headers":     map[string]string{"Content-Type": "text/html"},
		"body":        sampleHTML,
	}

	ctx.Set("http_result", httpResponse)

	config := map[string]interface{}{
		"html_source": "http_result",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)
	assert.Equal(t, "Welcome", results[0]["title"])
}

func TestHTMLParserTask_NonStringHTMLContent(t *testing.T) {
	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	// Set non-string content
	ctx.Set("html_content", 12345)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "title",
				"selector": "h1",
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "failed", result.Status)
	assert.Contains(t, result.Error, "not a string")
}

func TestHTMLParserTask_ComplexHTMLStructure(t *testing.T) {
	complexHTML := `
	<div class="products">
		<div class="product">
			<h2 class="title">Product 1</h2>
			<span class="price">$99.99</span>
			<a href="/product1">View</a>
		</div>
		<div class="product">
			<h2 class="title">Product 2</h2>
			<span class="price">$149.99</span>
			<a href="/product2">View</a>
		</div>
	</div>
	`

	task := &HTMLParserTask{}
	ctx := engine.NewExecutionContext()

	ctx.Set("html_content", complexHTML)

	config := map[string]interface{}{
		"html_source": "html_content",
		"selectors": []interface{}{
			map[string]interface{}{
				"name":     "titles",
				"selector": ".product .title",
				"multiple": true,
			},
			map[string]interface{}{
				"name":     "prices",
				"selector": ".product .price",
				"multiple": true,
			},
			map[string]interface{}{
				"name":      "links",
				"selector":  ".product a",
				"attribute": "href",
				"multiple":  true,
			},
		},
	}

	result := task.Execute(ctx, config)

	assert.Equal(t, "success", result.Status)

	results := result.Output.([]map[string]any)

	titles := results[0]["titles"].([]string)
	assert.Len(t, titles, 2)
	assert.Equal(t, "Product 1", titles[0])
	assert.Equal(t, "Product 2", titles[1])

	prices := results[0]["prices"].([]string)
	assert.Len(t, prices, 2)
	assert.Equal(t, "$99.99", prices[0])
	assert.Equal(t, "$149.99", prices[1])

	links := results[0]["links"].([]string)
	assert.Len(t, links, 2)
	assert.Equal(t, "/product1", links[0])
	assert.Equal(t, "/product2", links[1])
}

func TestHTMLParserTask_RegisterHTMLParserTask(t *testing.T) {
	registry := engine.NewRegistry()

	RegisterHTMLParserTask(registry)

	executor, err := registry.Get("html_parser")
	assert.NoError(t, err)
	assert.NotNil(t, executor)
	assert.IsType(t, &HTMLParserTask{}, executor)
}
