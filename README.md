# Monclissh Dashboard CLI Tool

Monclissh is a command-line interface (CLI) tool designed to monitor and display system metrics such as disk space, CPU load, and RAM usage for multiple configurable SSH hosts. This tool provides a dashboard that allows users to easily visualize the performance of their remote systems.

## Features

- Monitor multiple SSH hosts
- Display real-time metrics for CPU, memory, and disk usage
- Configurable through a YAML file
- Easy to use command-line interface

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/monclissh.git
   cd monclissh
   ```

2. Build the application:
   ```
   go build -o monclissh ./cmd/monclissh
   ```

3. Ensure you have Go installed on your machine. You can download it from [golang.org](https://golang.org/dl/).

## Configuration

The application requires a configuration file named `hosts.yaml` located in the `configs` directory. This file should contain the SSH hosts you want to monitor. An example configuration is shown below:

```yaml
hosts:
  - name: "Host1"
    ip: "192.168.1.1"
    user: "username"
    password: "password"
  - name: "Host2"
    ip: "192.168.1.2"
    user: "username"
    password: "password"
```

## Usage

To run the application, execute the following command:

```
./monclissh
```

The dashboard will start and display the metrics for the configured hosts. The metrics will be updated periodically.

## Metrics Displayed

- **CPU Load**: Current CPU usage percentage.
- **Memory Usage**: Amount of RAM used and available.
- **Disk Space**: Total and available disk space on the host.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.