# Microsoft Copilot Integration with Fabric

This document describes the Microsoft Copilot vendor implementation for Fabric, enabling users to access Microsoft 365 Copilot capabilities through the Fabric CLI.

## Overview

The Microsoft Copilot integration allows Fabric to communicate with Microsoft 365 Copilot using the official Microsoft Graph API. This integration provides:

- **Enterprise-grade AI** with Microsoft 365 data grounding
- **Secure authentication** using Microsoft Entra ID (Azure AD)
- **Conversation management** compatible with Fabric's stateless architecture
- **Full Fabric pattern support** for all existing prompts

## Prerequisites

### Microsoft 365 Requirements

1. **Microsoft 365 Copilot License** - Required for each user
2. **Microsoft Entra App Registration** - API access configuration
3. **Required Permissions** - Microsoft Graph API permissions:
   - `Sites.Read.All`
   - `Mail.Read`
   - `People.Read.All`
   - `OnlineMeetingTranscript.Read.All`
   - `Chat.Read`
   - `ChannelMessage.Read.All`
   - `ExternalItem.Read.All`

### App Registration Setup

1. Go to [Microsoft Entra admin center](https://entra.microsoft.com/)
2. Navigate to **App registrations** → **New registration**
3. Set:
   - **Name**: "Fabric Copilot Integration"
   - **Supported account types**: "Accounts in this organizational directory only"
   - **Redirect URI**: (not required for app-only access)
4. Go to **API permissions** → **Add a permission**
5. Add the required Microsoft Graph permissions listed above
6. Grant admin consent for the permissions
7. Go to **Certificates & secrets** → **New client secret**
8. Create a client secret and copy the value

## Configuration

### Environment Variables

Configure Fabric with the following environment variables:

```bash
# Required for app-only authentication (recommended)
export MICROSOFT_COPILOT_CLIENT_ID="your-app-client-id"
export MICROSOFT_COPILOT_CLIENT_SECRET="your-client-secret"

# Optional: Specify tenant (defaults to "common")
export MICROSOFT_COPILOT_TENANT_ID="your-tenant-id"

# Optional: Use OAuth flow instead of client secrets
export MICROSOFT_COPILOT_USE_OAUTH="true"
```

### Interactive Setup

Run Fabric setup and follow the prompts:

```bash
fabric --setup
```

The setup will ask for:

- **Client ID** (required)
- **Client Secret** (required for app-only access)
- **Tenant ID** (optional)
- **Use OAuth flow** (optional, defaults to false)

## Usage

### Basic Usage

```bash
# Use with default model
echo "What meetings do I have today?" | fabric --pattern summarize --vendor "Microsoft Copilot"

# Specify model explicitly
echo "Summarize this document" | fabric --pattern extract_wisdom --vendor "Microsoft Copilot" --model copilot-enterprise
```

### Pattern Examples

Microsoft Copilot works with all existing Fabric patterns:

```bash
# Summarize content
cat document.txt | fabric --pattern summarize --vendor "Microsoft Copilot"

# Extract wisdom
cat article.md | fabric --pattern extract_wisdom --vendor "Microsoft Copilot"

# Analyze answers
cat interview.txt | fabric --pattern analyze_answers --vendor "Microsoft Copilot"
```

## Features

### Supported Capabilities

✅ **Chat Completions** - Full conversation support
✅ **Pattern Integration** - Works with all Fabric patterns
✅ **Enterprise Search** - Grounded in Microsoft 365 data
✅ **Secure Authentication** - Microsoft Entra ID integration
✅ **Error Handling** - Robust error recovery
✅ **Session Management** - Automatic conversation lifecycle

### Current Limitations

⚠️ **Streaming** - Currently uses non-streaming response (will be enhanced)
⚠️ **OAuth Flow** - Basic implementation, client secret recommended
⚠️ **File Context** - Not yet implemented (planned)
⚠️ **Web Search Control** - Uses default Copilot settings

## Architecture

### Conversation Management

Microsoft Copilot uses a stateful conversation API, while Fabric expects stateless calls. The implementation bridges this gap by:

1. **Creating conversations** for each Fabric request
2. **Managing lifecycle** automatically (create → use → cleanup)
3. **Converting messages** between Fabric and Copilot formats
4. **Extracting responses** from complex conversation objects

### Authentication Flow

The implementation supports two authentication methods:

1. **App-Only Access** (Recommended)
   - Uses client credentials flow
   - No user interaction required
   - Suitable for CLI/automation

2. **OAuth Flow** (Basic)
   - Interactive user authentication
   - Requires browser-based login
   - Stores tokens securely

## Troubleshooting

### Common Issues

#### Authentication Errors

```text
Error: failed to create conversation, status: 401
```

- Verify client ID and secret are correct
- Check app registration permissions
- Ensure admin consent is granted

#### Permission Errors

```text
Error: failed to create conversation, status: 403
```

- Verify all required permissions are granted
- Check tenant ID configuration
- Ensure user has Copilot license

#### Network Issues

```text
Error: failed to send token request: connection timeout
```

- Check network connectivity
- Verify firewall settings
- Try again later

### Debug Mode

Enable debug logging to troubleshoot:

```bash
fabric --vendor "Microsoft Copilot" --pattern summarize --debug
```

## Development

### Running Tests

```bash
# Run all tests
go test ./internal/plugins/ai/copilot -v

# Run specific test
go test ./internal/plugins/ai/copilot -run TestNewClient -v
```

### Project Structure

```text
internal/plugins/ai/copilot/
├── copilot.go          # Main vendor implementation
├── auth.go             # Authentication logic
├── conversation.go      # Conversation management
├── models.go           # API data structures
├── copilot_test.go     # Test suite
```

## Security Considerations

- **Tokens** are stored securely in `~/.config/fabric/.copilot_oauth`
- **Client Secrets** should be treated as sensitive credentials
- **Permissions** follow principle of least privilege
- **Data Access** respects Microsoft 365 security boundaries

## Contributing

To enhance the Microsoft Copilot integration:

1. **Implement streaming** response support
2. **Add file context** handling
3. **Enhance OAuth** flow with PKCE
4. **Add web search** control options
5. **Improve error** messages and recovery

See the [Fabric contributing guidelines](../../../../docs/CONTRIBUTING.md) for development setup.

## License

This integration follows the same license as Fabric project.
