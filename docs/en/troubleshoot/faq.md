# FAQ

## Basics & Installation

### Q: What is Podux? How is it different from plain frpc?

A: Podux is a graphical management tool built on top of frpc. It integrates the core frp client functionality and provides a modern web interface for managing configuration, monitoring status, and viewing logs — eliminating the need to manually edit config files and the lack of visual monitoring in vanilla frpc.

### Q: Which operating systems does Podux support?

A: Windows, macOS, and major Linux distributions are all supported.

## Configuration & Usage

### Q: Can I import my existing `frpc.toml` config file?

A: It is currently recommended to recreate your configuration through Podux's visual interface to ensure all fields are correctly parsed and managed.

### Q: How do configuration changes take effect?

A: Podux supports hot reload. After saving changes in the web UI, the application automatically applies the new configuration — a full service restart is usually not required.

### Q: Which proxy protocols are supported?

A: **TCP**, **UDP**, **HTTP**, and **HTTPS** are fully supported. Advanced protocols such as **STCP** and **XTCP** are under development and will be available in future releases.

## Troubleshooting

### Q: How do I report a bug or unknown error?

A: Please open an issue on the GitHub repository. Include detailed error logs, steps to reproduce, and your operating system information to help us diagnose the problem quickly.
