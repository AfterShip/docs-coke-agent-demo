#!/bin/bash

# Script to quickly add a new tool
# Usage: ./add_tool.sh tool_name "Tool Description"

if [ $# -ne 2 ]; then
    echo "Usage: $0 <tool_name> <tool_description>"
    echo "Example: $0 calculator 'Evaluate mathematical expressions'"
    exit 1
fi

TOOL_NAME=$1
TOOL_DESC=$2
TOOL_FILE="${TOOL_NAME}.go"

# Convert tool_name to CamelCase for struct names
TOOL_STRUCT=$(echo "${TOOL_NAME}" | sed 's/_\(.\)/\U\1/g' | sed 's/^./\U&/')

echo "Creating new tool: ${TOOL_NAME}"
echo "Description: ${TOOL_DESC}"
echo "File: ${TOOL_FILE}"
echo "Struct: ${TOOL_STRUCT}Tool"

# Create the tool file from template
sed "s/ExampleTool/${TOOL_STRUCT}Tool/g; s/example_tool/${TOOL_NAME}/g; s/ExampleRequest/${TOOL_STRUCT}Request/g; s/ExampleResponse/${TOOL_STRUCT}Response/g; s/Description of what this tool does/${TOOL_DESC}/g" example_new_tool.go.template > "${TOOL_FILE}"

echo "✅ Created ${TOOL_FILE}"

# Add to registry
if ! grep -q "NewExample" interface.go; then
    echo "⚠️  Note: You need to manually add registry.Register(New${TOOL_STRUCT}Tool()) to GetDefaultRegistry() in interface.go"
else
    echo "📝 Remember to update GetDefaultRegistry() in interface.go to register your new tool"
fi

echo ""
echo "Next steps:"
echo "1. Edit ${TOOL_FILE} and implement your tool logic"
echo "2. Add registry.Register(New${TOOL_STRUCT}Tool()) to GetDefaultRegistry() in interface.go"
echo "3. Test your tool with: go test ./apps/listingagent/internal/domain/tool -v"