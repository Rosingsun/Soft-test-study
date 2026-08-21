.PHONY: test test-verbose cover build vet

# 运行全部测试
test:
	go test ./...

# 详细输出
test-verbose:
	go test ./internal/service/... ./pkg/jwt/... ./pkg/security/... -v

# 覆盖率报告（service + pkg）
cover:
	go test -cover ./internal/service/... ./pkg/jwt/... ./pkg/security/...

# 编译检查
build:
	go build ./...

# 静态检查
vet:
	go vet ./...
