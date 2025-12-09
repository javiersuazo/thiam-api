#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="$(dirname "$SCRIPT_DIR")"
SPEC_FILE="$API_DIR/openapi.yaml"

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Generate TypeScript client from OpenAPI specification.

OPTIONS:
    -o, --output DIR     Output directory for generated files (default: ./generated)
    -t, --type TYPE      Output type: typescript, fetch, axios (default: typescript)
    -m, --mock           Start mock server instead of generating client
    -p, --port PORT      Mock server port (default: 4010)
    -h, --help           Show this help message

EXAMPLES:
    # Generate TypeScript types
    $(basename "$0") -o ./src/types

    # Generate fetch client
    $(basename "$0") -t fetch -o ./src/api

    # Start mock server for development
    $(basename "$0") --mock

    # Start mock server on custom port
    $(basename "$0") --mock -p 8081
EOF
}

OUTPUT_DIR="./generated"
OUTPUT_TYPE="typescript"
MOCK_MODE=false
MOCK_PORT=4010

while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -t|--type)
            OUTPUT_TYPE="$2"
            shift 2
            ;;
        -m|--mock)
            MOCK_MODE=true
            shift
            ;;
        -p|--port)
            MOCK_PORT="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

if [ ! -f "$SPEC_FILE" ]; then
    echo "Error: OpenAPI spec not found at $SPEC_FILE"
    exit 1
fi

if $MOCK_MODE; then
    echo "Starting mock server on port $MOCK_PORT..."
    echo "API will be available at http://localhost:$MOCK_PORT"
    echo ""
    echo "Press Ctrl+C to stop"
    npx @stoplight/prism-cli mock "$SPEC_FILE" --port "$MOCK_PORT" --host 0.0.0.0
else
    mkdir -p "$OUTPUT_DIR"

    case $OUTPUT_TYPE in
        typescript)
            echo "Generating TypeScript types..."
            npx openapi-typescript "$SPEC_FILE" -o "$OUTPUT_DIR/api.ts"
            echo "Generated: $OUTPUT_DIR/api.ts"
            ;;
        fetch)
            echo "Generating TypeScript types for use with openapi-fetch..."
            npx openapi-typescript "$SPEC_FILE" -o "$OUTPUT_DIR/api.ts"
            echo "Generated: $OUTPUT_DIR/api.ts"
            echo ""
            echo "To use with openapi-fetch, install it in your project:"
            echo "  npm install openapi-fetch"
            echo ""
            echo "Then use in your code:"
            echo "  import createClient from 'openapi-fetch'"
            echo "  import type { paths } from '$OUTPUT_DIR/api'"
            echo "  const client = createClient<paths>({ baseUrl: 'http://localhost:8080' })"
            ;;
        axios)
            echo "Generating axios client..."
            npx @openapitools/openapi-generator-cli generate \
                -i "$SPEC_FILE" \
                -g typescript-axios \
                -o "$OUTPUT_DIR"
            echo "Generated axios client in: $OUTPUT_DIR"
            ;;
        *)
            echo "Unknown output type: $OUTPUT_TYPE"
            echo "Supported types: typescript, fetch, axios"
            exit 1
            ;;
    esac

    if [ "$OUTPUT_TYPE" = "typescript" ]; then
        echo ""
        echo "Done! Import in your code:"
        echo "  import type { paths, components } from '$OUTPUT_DIR/api.ts'"
    fi
fi
