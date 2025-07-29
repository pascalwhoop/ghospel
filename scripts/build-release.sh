#!/bin/bash

# Ghospel Release Build Script
# Builds whisper.cpp binaries for multiple architectures and creates release builds

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
WHISPER_DIR="$ROOT_DIR/whisper_cpp_source"
BINARIES_DIR="$ROOT_DIR/internal/binaries"

echo "🚀 Starting Ghospel release build process..."

# Ensure we're in the right directory
cd "$ROOT_DIR"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf "$BINARIES_DIR"
mkdir -p "$BINARIES_DIR"

# Function to build whisper.cpp for a specific platform/arch
build_whisper() {
    local platform=$1
    local arch=$2
    local cross_compile=${3:-""}
    
    echo "🏗️  Building whisper.cpp for $platform-$arch..."
    
    cd "$WHISPER_DIR"
    
    # Clean previous build
    rm -rf build
    
    local cmake_flags=""
    local build_dir="build-$platform-$arch"
    
    case "$platform" in
        "darwin")
            case "$arch" in
                "arm64")
                    cmake_flags="-DCMAKE_OSX_ARCHITECTURES=arm64 -DGGML_METAL=ON -DGGML_METAL_EMBED_LIBRARY=ON -DGGML_BLAS_DEFAULT=ON"
                    ;;
                "amd64")
                    cmake_flags="-DCMAKE_OSX_ARCHITECTURES=x86_64 -DGGML_METAL=ON -DGGML_METAL_EMBED_LIBRARY=ON -DGGML_BLAS_DEFAULT=ON"
                    ;;
            esac
            ;;
        "linux")
            case "$arch" in
                "arm64")
                    cmake_flags="-DCMAKE_SYSTEM_PROCESSOR=aarch64 -DGGML_BLAS_DEFAULT=ON"
                    if [ "$cross_compile" = "true" ]; then
                        cmake_flags="$cmake_flags -DCMAKE_C_COMPILER=aarch64-linux-gnu-gcc -DCMAKE_CXX_COMPILER=aarch64-linux-gnu-g++"
                    fi
                    ;;
                "amd64")
                    cmake_flags="-DGGML_BLAS_DEFAULT=ON"
                    ;;
            esac
            ;;
    esac
    
    # Build
    cmake -B "$build_dir" $cmake_flags \
        -DCMAKE_BUILD_TYPE=Release \
        -DWHISPER_BUILD_TESTS=OFF \
        -DWHISPER_BUILD_SERVER=OFF
    
    cmake --build "$build_dir" -j$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4) --config Release
    
    # Copy binary to binaries directory
    local binary_name="whisper-cli-$platform-$arch"
    local binary_path="$build_dir/bin/whisper-cli"
    
    if [ -f "$binary_path" ]; then
        cp "$binary_path" "$BINARIES_DIR/$binary_name"
        echo "✅ Built $binary_name successfully"
    else
        echo "❌ Failed to build $binary_name"
        return 1
    fi
    
    cd "$ROOT_DIR"
}

# Detect current platform
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)
if [ "$CURRENT_ARCH" = "x86_64" ]; then
    CURRENT_ARCH="amd64"
fi

echo "📍 Current platform: $CURRENT_OS-$CURRENT_ARCH"

# Build for current platform first
build_whisper "$CURRENT_OS" "$CURRENT_ARCH"

# Build for other architectures if on macOS (can cross-compile easily)
if [ "$CURRENT_OS" = "darwin" ]; then
    if [ "$CURRENT_ARCH" = "arm64" ]; then
        echo "🔄 Cross-compiling for Intel Mac..."
        build_whisper "darwin" "amd64"
    else
        echo "🔄 Cross-compiling for Apple Silicon..."
        build_whisper "darwin" "arm64"
    fi
fi

# Note: Linux cross-compilation requires additional setup
# Users would need: sudo apt-get install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
# if [ "$CURRENT_OS" = "linux" ] && [ "$CURRENT_ARCH" = "amd64" ]; then
#     if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then
#         echo "🔄 Cross-compiling for Linux ARM64..."
#         build_whisper "linux" "arm64" "true"
#     else
#         echo "⚠️  Skipping Linux ARM64 build (cross-compiler not available)"
#     fi
# fi

echo "📦 Built binaries:"
ls -la "$BINARIES_DIR/"

echo "🏗️  Building Go release binary..."
go build -tags release -ldflags "-s -w" -o ghospel ./cmd/ghospel

echo "✅ Release build complete!"
echo ""
echo "📋 Summary:"
echo "  - Whisper binaries: $BINARIES_DIR/"
echo "  - Go binary: ./ghospel"
echo "  - Ready for distribution!"