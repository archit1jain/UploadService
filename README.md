# File Transfer

A minimal web application for transferring files from a mobile phone to a computer over a local Wi-Fi network.

## Requirements

- Go (1.16+ recommended)

## How to Run

1. Navigate to the directory containing this project.
2. Run the application:
   ```bash
   go run .
   ```

By default, the server listens on port `8080` and saves uploaded files to `./uploads`.

## Configuration

You can configure the application using environment variables:

- `PORT`: The port the server should listen on (default: `8080`).
- `UPLOAD_DIR`: The directory where uploaded files will be saved (default: `./uploads`).

Example:
```bash
PORT=9000 UPLOAD_DIR=/path/to/my/folder go run .
```

## How to Connect from a Mobile Phone

1. Ensure both your computer and your mobile phone are connected to the same Wi-Fi network.
2. Find your computer's local IP address:
   - **macOS/Linux**: Run `ifconfig` or `ip a` in the terminal and look for `en0`, `eth0`, or `wlan0`. It usually looks like `192.168.1.x` or `10.0.0.x`.
   - **Windows**: Run `ipconfig` in the Command Prompt and look for "IPv4 Address".
3. Open a web browser on your phone and navigate to:
   ```
   http://<your-computer-local-ip>:8080
   ```
   *(Replace `8080` if you configured a different `PORT`)*

## Troubleshooting

- **Page not loading on phone**: Ensure your computer's firewall is not blocking incoming connections on the configured port (`8080`). You may need to add a firewall rule to allow traffic to the Go executable or the specific port.
- **Upload fails**: Ensure the application has write permissions to the `UPLOAD_DIR`.
