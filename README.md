## Parody
"Yamcha is a basic HTTP load testing tool that works (Power lvl : 1480), showcasing mediocrity at its finest. Woah there, don't expect too much of him. He often ends up like ⬇️

![FkxqjP1aYAAyT5X](https://github.com/Aditya1404Sal/Yamcha/assets/91340059/74915949-c768-401e-acdf-4d581c468725)

---
# Yamcha : A Load Testing Tool

Yamcha is a command-line load testing tool written in Go for conducting performance tests on HTTP/S applications.

## Features

- **CLI-based**: Easy-to-use command-line interface for running load tests.
- **Configurable HTTP Methods**: Supports HTTP methods such as GET and POST for different testing scenarios.
- **Concurrent Request Handling**: Utilizes goroutines (number depends on allowed cpu cores) for concurrent request handling.
- **CPU Utilization Control**: Allows setting the number of CPUs to utilize during load tests.
- **Optional Plotting**: Integrates optional plotting functionality to visualize load test results.

## Installation

Clone the repository:

```bash
git clone https://github.com/your-username/yamcha-load-testing-tool.git
cd yamcha-load-testing-tool
```

Build the executable:

```bash
go build -o yamcha
```

## Usage

Run a basic load test:

```bash
./yamcha --url=https://example.com --requests=100 --rate=10 --method=GET
```

### Command-line Flags

- `-url`: Specify the URL of the target application.
- `-requests`: Number of requests to send during the test.
- `-rate`: Number of requests per second to send.
- `-method`: HTTP method to use (GET, POST, etc.).
- `-cpus`: Number of CPUs to utilize (optional).
- `-plot`: Enable plotting of load test results (optional).
