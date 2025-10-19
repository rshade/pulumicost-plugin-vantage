# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the
PulumiCost Vantage plugin.

## Authentication Errors

### "401 Unauthorized" or "Invalid token"

**Symptoms:**

- API requests fail with authentication errors
- Error messages mentioning invalid or expired tokens

**Solutions:**

1. Verify your API token is valid and not expired
2. Check that `PULUMICOST_VANTAGE_TOKEN` environment variable is set correctly
3. Ensure you're using the correct token type:
   - Cost Report tokens for report-specific access
   - Workspace tokens for broader workspace access
4. Regenerate your token in Vantage if it has expired

**Example:**

```bash
# Check if token is set
echo $PULUMICOST_VANTAGE_TOKEN

# Test with a simple API call (replace with your token)
curl -H "Authorization: Bearer YOUR_TOKEN" https://api.vantage.sh/v1/costs?cost_report_token=cr_xxx&start_at=2024-01-01&end_at=2024-01-02
```

## Rate Limiting (429 Errors)

### "429 Too Many Requests"

**Symptoms:**

- Requests fail with rate limit errors
- Operations take longer than expected due to retries

**Solutions:**

1. The plugin automatically handles rate limits with exponential backoff
2. Check `X-RateLimit-*` headers in Vantage responses for limit information
3. Reduce request frequency or increase retry intervals
4. Consider using Cost Report tokens instead of Workspace tokens for
   better rate limits

**Configuration:**

```yaml
params:
  max_retries: 10  # Increase retry attempts
  request_timeout_seconds: 120  # Allow more time for retries
```

## Pagination Issues

### "Cursor expired" or pagination errors

**Symptoms:**

- Large data syncs fail midway through pagination
- Inconsistent results across runs

**Solutions:**

1. Reduce `page_size` to decrease memory usage and timeout risks
2. Increase `request_timeout_seconds` for large pages
3. Use smaller date ranges for backfills
4. Enable verbose logging to inspect pagination cursors

**Configuration:**

```yaml
params:
  page_size: 1000  # Reduce from default 5000
  request_timeout_seconds: 120
```

## Data Sync Issues

### Missing or incomplete data

**Symptoms:**

- Expected cost records are missing
- Inconsistent data between runs

**Solutions:**

1. Check date ranges in configuration
2. Verify group_by dimensions are available in your Vantage setup
3. Use bookmarks for incremental syncs to avoid duplicates
4. Enable verbose logging to inspect raw API responses

### Large backfills failing

**Symptoms:**

- Backfill operations timeout or run out of memory
- Operations take longer than expected

**Solutions:**

1. Break large backfills into smaller date ranges
2. Use CSV export for very large historical datasets
3. Increase system resources or run during off-peak hours
4. Monitor memory usage during operations

**Example:**

```bash
# Instead of 12 months at once
./bin/pulumicost-vantage backfill --config config.yaml --months 3
./bin/pulumicost-vantage backfill --config config.yaml --months 3 --start-date 2024-04-01
./bin/pulumicost-vantage backfill --config config.yaml --months 3 --start-date 2024-07-01
```

## Verbose Logging

### Enabling detailed logs

To troubleshoot issues, enable verbose logging to see raw API requests and responses:

**Environment Variables:**

```bash
export PULUMICOST_VANTAGE_LOG_LEVEL=debug
export PULUMICOST_VANTAGE_VERBOSE=true
```

**Configuration:**

```yaml
params:
  verbose_logging: true
  log_raw_responses: true
```

**Log Output:**

- Structured logs with `adapter=vantage` field
- Request/response details (with sensitive data redacted)
- Pagination cursor information
- Retry attempts and backoff timing

## Mock Server Testing

### Using Wiremock for local testing

The plugin includes a mock server for testing without hitting live APIs:

**Start mock server:**

```bash
make wiremock-up
```

**Run tests against mock:**

```bash
make demo
```

**Capture real API recordings:**

1. Start Wiremock in record mode
2. Run plugin against live API
3. Save recordings for future tests

**Example mock configuration:**

```yaml
# test/config-mock.yaml
version: 0.1
source: vantage
credentials:
  token: mock_token
params:
  cost_report_token: "cr_mock"
  start_date: "2024-01-01"
  end_date: "2024-01-02"
```

## Common Configuration Issues

### Invalid date formats

**Error:** Date parsing failures
**Solution:** Use ISO format (YYYY-MM-DD) for all dates

### Missing required fields

**Error:** Configuration validation failures
**Solution:** Ensure all required fields are present (see CONFIG.md)

### Token scope issues

**Error:** Access denied to specific resources
**Solution:** Verify token has appropriate permissions for the requested operations

## Performance Tuning

### Slow syncs

**Symptoms:**

- Operations take longer than expected
- High memory usage

**Solutions:**

1. Reduce page size
2. Use appropriate granularity (day vs month)
3. Limit group_by dimensions
4. Enable incremental sync with bookmarks

**Performance configuration:**

```yaml
params:
  page_size: 1000
  granularity: "month"  # For historical data
  group_bys: ["provider", "service", "account"]  # Limit dimensions
```

## Getting Help

### Debug information to collect

When reporting issues, include:

1. Plugin version and Go version
2. Configuration (with tokens redacted)
3. Full error logs with verbose mode enabled
4. Vantage workspace/report details
5. Approximate data volume being processed

### Log analysis

Look for these patterns in logs:

- `rate_limit_exceeded` - Rate limiting issues
- `cursor_expired` - Pagination problems
- `auth_failed` - Authentication issues

## Emergency Procedures

### Stopping runaway processes

If a sync operation becomes unresponsive:

```bash
# Find the process
ps aux | grep pulumicost-vantage

# Terminate gracefully
kill -TERM <pid>

# Force kill if needed
kill -KILL <pid>
```

### Data recovery

For incomplete syncs:

1. Check bookmark state
2. Resume from last successful sync point
3. Use smaller batch sizes for retry
