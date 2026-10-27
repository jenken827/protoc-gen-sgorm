#!/usr/bin/bash
# protoc-gen-sgorm 开发脚本：生成、测试、格式化都走这里，不要直接跑 go / buf。
set -euo pipefail

cd "$(cd "$(dirname "$0")" && pwd)"
BIN="$PWD/bin"
export PATH="$BIN:$PATH"

usage() {
    cat <<'EOF'
用法: dev.sh <命令>

命令:
  gen      编译插件并重新生成 options.pb.go / example 下的代码
  test     gen + go vet + go build + go test + 编译生成的示例实体
  fmt      gofmt 整理 Go 源码
  clean    删掉 bin/ 与生成物
  help     显示本帮助
EOF
}

need() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "❌ 需要 $1，但 PATH 里找不到" >&2
        exit 1
    }
}

cmd_gen() {
    need go
    need buf
    mkdir -p "$BIN"

    # 顺序有讲究：插件自身 import 了 gen/sgorm（options 的 pb），
    # 所以必须先用 protoc-gen-go 生成 options.pb.go，再编译插件。
    echo "🔧 准备依赖并编译 protoc-gen-go..."
    go mod download google.golang.org/protobuf
    go build -o "$BIN/protoc-gen-go" google.golang.org/protobuf/cmd/protoc-gen-go

    echo "📦 生成 options / 示例的 Go 代码..."
    buf generate --template buf.gen.pb.yaml

    echo "🔧 编译 protoc-gen-sgorm..."
    go build -o "$BIN/protoc-gen-sgorm" ./cmd/protoc-gen-sgorm

    echo "📦 生成示例实体..."
    rm -f example/gen/entity/*.sgorm.go
    buf generate
    go mod tidy

    echo "✅ 生成完成"
}

cmd_test() {
    cmd_gen
    echo "🧪 go vet / build / test..."
    go vet ./...
    go build ./...
    go test ./...
    go build ./example/gen/entity
    echo "✅ 全部通过"
}

cmd_fmt() {
    gofmt -l -w cmd internal example/*.go
}

cmd_clean() {
    rm -rf "$BIN"
    rm -f gen/sgorm/*.pb.go example/gen/native/*.pb.go example/gen/entity/*.sgorm.go
    echo "已清理生成物"
}

case "${1:-help}" in
gen) cmd_gen ;;
test) cmd_test ;;
fmt) cmd_fmt ;;
clean) cmd_clean ;;
help | -h | --help) usage ;;
*)
    echo "未知命令: $1" >&2
    usage
    exit 1
    ;;
esac
