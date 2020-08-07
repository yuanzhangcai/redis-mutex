default: sonar

# sonar代码扫描
sonar:
	ulimit -n 24000
	go vet -n ./... 2> ./vet.tmp
	golangci-lint run ./... --out-format=checkstyle > golangci-lint.tmp || true
	go test -race -cover -v  ./... -json -coverprofile=covprofile > test.tmp
	sonar-scanner \
	-Dsonar.host.url=http://127.0.0.1:9000 \
	-Dsonar.sources=. \
	-Dsonar.tests=. \
	-Dsonar.exclusions="**/*_test.go,**/examples/**" \
	-Dsonar.projectKey=redis-mutex \
	-Dsonar.login=b0bb49716ff66b69d44b9fb034a2dde14f9fb59b \
	-Dsonar.go.tests.reportPaths=test.tmp \
	-Dsonar.go.coverage.reportPaths=covprofile \
	-Dsonar.go.govet.reportPaths=vet.tmp \
	-Dsonar.go.golangci-lint.reportPaths=golangci-lint.tmp \
	-Dsonar.test.inclusions="**/*_test.go" \
    -Dsonar.test.exclusions="**/vendor/** " | grep -v "WARN:"
	rm -rf *.tmp
	rm -rf .scannerwork
	rm -rf covprofile


.PHONY: sonar