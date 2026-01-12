# Logging Setup Guide

## Overview

The JourneyBuilder application now supports **dual logging**: logs are written to both the console (stderr) and optionally to a log file.

## Features

- ✅ **Console Logging**: Always enabled (writes to stderr)
- ✅ **File Logging**: Optional, configurable via environment variable
- ✅ **Automatic Directory Creation**: Creates log directory if it doesn't exist
- ✅ **Append Mode**: Logs are appended to existing files (no overwrite)
- ✅ **Timestamped Logs**: All logs include date, time, and file location
- ✅ **Backward Compatible**: Existing `log` package calls work without changes

## Configuration

### Environment Variable

Set the `LOG_FILE` environment variable to enable file logging:

```bash
# In .env file
LOG_FILE=logs/app.log

# Or export in shell
export LOG_FILE=logs/app.log
```

### Log File Paths

You can use any path:

```bash
# Relative path (creates logs/ directory)
LOG_FILE=logs/app.log

# Absolute path
LOG_FILE=/var/log/journey-builder/app.log

# Nested directories (auto-created)
LOG_FILE=logs/production/app.log
```

### Disable File Logging

If `LOG_FILE` is not set or empty, logs only go to console:

```bash
# No file logging (console only)
# Don't set LOG_FILE, or set it to empty
unset LOG_FILE
```

## Usage

### Automatic Initialization

The logger is automatically initialized in `main.go` when the application starts. No code changes needed in other files - all existing `log.Println()`, `log.Printf()`, etc. calls will automatically write to both console and file (if configured).

### Example Log Output

**Console Output:**
```
2024/01/15 10:30:45 main.go:25: Info: .env file not found, using system environment variables
2024/01/15 10:30:45 main.go:28: ✓ Loaded .env file
2024/01/15 10:30:45 logger.go:45: ✓ Logging to file: logs/app.log
2024/01/15 10:30:45 main.go:40: Initializing AI services...
```

**Log File** (`logs/app.log`):
```
2024/01/15 10:30:45 main.go:25: Info: .env file not found, using system environment variables
2024/01/15 10:30:45 main.go:28: ✓ Loaded .env file
2024/01/15 10:30:45 logger.go:45: ✓ Logging to file: logs/app.log
2024/01/15 10:30:45 main.go:40: Initializing AI services...
2024/01/15 10:30:46 gemini.go:72: ✓ Gemini client created
2024/01/15 10:30:46 main.go:92: 🚀 Server starting on port 8080
```

## Log Format

Each log entry includes:
- **Date & Time**: `2024/01/15 10:30:45`
- **File Location**: `main.go:25` (file name and line number)
- **Message**: The actual log message

Format: `YYYY/MM/DD HH:MM:SS filename.go:line: message`

## Implementation Details

### Logger Package

The logger is implemented in `internal/logger/logger.go`:

- Uses `io.MultiWriter` to write to multiple destinations simultaneously
- Creates log directories automatically if they don't exist
- Opens log files in append mode (preserves existing logs)
- Gracefully handles errors (falls back to console-only if file logging fails)

### Integration

The logger is initialized early in `main.go`:

```go
// Initialize logger (file + console logging)
if err := logger.InitLogger(); err != nil {
    log.Printf("Warning: Failed to initialize logger: %v. Continuing with console logging only.", err)
} else {
    logFile := os.Getenv("LOG_FILE")
    if logFile != "" {
        log.Printf("✓ Logging to file: %s", logFile)
    }
}
defer logger.Close()  // Cleanup on exit
```

## Best Practices

### 1. Log Rotation

The logger doesn't handle log rotation automatically. For production, consider:

- **External Tools**: Use `logrotate` (Linux) or similar tools
- **Size Limits**: Monitor log file size and rotate manually
- **Date-based Files**: Use date in filename: `LOG_FILE=logs/app-$(date +%Y-%m-%d).log`

### 2. Log File Permissions

Log files are created with `0666` permissions (readable/writable by all). For production:

- Set appropriate file permissions after creation
- Use a dedicated user/group for the application
- Consider using `umask` to restrict permissions

### 3. Disk Space

Monitor log file size:

```bash
# Check log file size
ls -lh logs/app.log

# Rotate when file gets too large
mv logs/app.log logs/app.log.old
touch logs/app.log
```

### 4. Development vs Production

**Development** (console only):
```bash
# .env
# LOG_FILE not set
```

**Production** (file + console):
```bash
# .env
LOG_FILE=/var/log/journey-builder/app.log
```

## Troubleshooting

### Log File Not Created

1. **Check permissions**: Ensure the application has write permissions
2. **Check path**: Verify the directory path is correct
3. **Check environment**: Ensure `LOG_FILE` is set correctly

```bash
# Verify environment variable
echo $LOG_FILE

# Check permissions
ls -ld logs/
```

### Logs Not Appearing in File

1. **Check initialization**: Look for "Logging to file" message in console
2. **Check errors**: Look for logger initialization errors
3. **Verify file path**: Check if file exists and is being written to

### File Permission Errors

If you see permission errors:

```bash
# Create directory with proper permissions
mkdir -p logs
chmod 755 logs

# Or use absolute path with proper permissions
LOG_FILE=/var/log/journey-builder/app.log
sudo mkdir -p /var/log/journey-builder
sudo chown $USER:$USER /var/log/journey-builder
```

## Example .env Configuration

```bash
# Application Configuration
PORT=8080
GEMINI_API_KEY=your_key_here
GEMINI_MODEL=gemini-2.5-flash

# Logging Configuration
LOG_FILE=logs/app.log
```

## Summary

- ✅ File logging is **optional** - set `LOG_FILE` to enable
- ✅ Console logging is **always enabled**
- ✅ All existing `log` package calls work automatically
- ✅ Logs include timestamps and file locations
- ✅ Log directories are created automatically
- ✅ Logs are appended (not overwritten)

---

*For more details on logging locations, see [LOGGING_LOCATIONS.md](./LOGGING_LOCATIONS.md)*
