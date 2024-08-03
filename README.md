## Parody
"Yamcha is a basic HTTP load testing tool that works (Power lvl : 1480), showcasing mediocrity at its finest. Woah there, don't expect too much of him. He often ends up like ⬇️

![FkxqjP1aYAAyT5X](https://github.com/Aditya1404Sal/Yamcha/assets/91340059/74915949-c768-401e-acdf-4d581c468725)

---
# Yamcha : A Load Testing Tool

Yamcha is a command-line load testing tool written in Go for conducting performance tests on HTTP/S applications.

## Features

- **CLI-based**: Easy-to-use command-line interface for running load tests.
- **Various Attack Patterns**: Supports Multiple attack variations like Steady, Burst, Spike, Random, Ramp-Up with more to come.
- **Configurable HTTP Methods**: Supports HTTP methods such as GET and POST for different testing scenarios.
- **Request Body and Headers**: Allows specifying request body and headers via a JSON file.
- **Concurrent Request Handling**: Utilizes goroutines (number depends on allowed cpu cores) for concurrent request handling.
- **CPU Utilization Control**: Allows setting the number of CPUs to utilize during load tests.
- **Optional Plotting**: Integrates optional plotting functionality to visualize load test results.
- **Live Test Progress**: Allows seamless visualization of test progress.

## Installation

Clone the repository:

```bash
git clone https://github.com/your-username/Yamcha.git
cd Yamcha
```

Build the executable:

```bash
go build -o yamcha
```

## Usage

Run a basic load test:

```bash
./yamcha -url https://example.com -requests 100 -rate 10 -method GET
```

Or Run a default test for `localhost:8080`

```bash
./yamcha
```

### Command-line Flags

- `-url`: Specify the URL of the target application.
- `-req`: Number of requests to send during the test.
- `-attack`: Type of attack to perform (steady, random, burst, rampup, spike, sustained).
- `-method`: HTTP method to use (GET, POST, etc.).
- `-rate`: Number of requests per second to send.
- `-burst`: Number of bursts for burst load attack.
- `-ss`: Step size for ramp up load attack.
- `-sh`: Spike Height for spike load attack.
- `-dur`: Duration for sustained load tests.
- `-cpu`: Number of CPUs to utilize (optional).
- `-plot`: Enable plotting of load test results as a time series (optional).
- `-body`: Path to a JSON file specifying request body and headers.
- `-conn`: Number of Active connections

### Test Status Bar
Enhance your load testing experience with a real-time progress bar, thanks to the `github.com/schollz/progressbar/v3 library`. The progress bar provides visual feedback on the test's progress, helping you to track the status of your load tests efficiently.

![Screenshot from 2024-07-25 22-37-56](https://github.com/user-attachments/assets/9845d172-4163-46fc-a899-e7659b459f16)
---
![Screenshot from 2024-07-25 22-38-11](https://github.com/user-attachments/assets/227ab384-ba22-41fe-a1f4-c1f55a8384f1)



### Example of body.json

```json
{
    "Headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer token",
        "Session-ID" : "Session ID here"
    },
    "Body": {
        "message": "Hello world",
        "user": "test_user_1",
        "timestamp": "2024-07-15"
    }
}

```
# Result Plot 
![Screenshot from 2024-07-20 22-04-00](https://github.com/user-attachments/assets/39ad42ed-92c3-4d68-a5ef-f17c467c842f)


