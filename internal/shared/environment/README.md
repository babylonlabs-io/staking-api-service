# Environment-Based Configuration

This package provides environment-specific configuration values for the staking API service. It allows hardcoding certain configuration values based on the deployment environment, which is determined by an environment variable.

## Features

1. **Environment Detection**: Automatically detects the environment based on the `BABYLON_ENVIRONMENT` environment variable.
2. **Hardcoded Values**: Provides hardcoded values for delegation transition parameters based on the environment.
3. **Config Integration**: Easily integrates with the existing config structure.

## Usage

### Environment Service

The `environment.Service` provides access to environment-specific configuration values:

```go
// Create a new environment service
envService := environment.NewService()

// Get the environment name
envName := envService.GetEnvironmentName()

// Get the delegation transition parameters
params := envService.GetDelegationTransitionParams()

// Apply the delegation transition parameters to the config
envService.ApplyDelegationTransitionParams(cfg)
```

### Environments

The following environments are supported:

- `mock-mainnet`
- `phase-2-devnet`
- `phase-2-testnet`
- `phase-2-private-mainnet`

### Delegation Transition Parameters

The delegation transition parameters are hardcoded for each environment:

| Environment | EligibleBeforeBtcHeight | AllowListExpirationHeight |
|-------------|-------------------------|---------------------------|
| mock-mainnet | 882251 | 8844 |
| phase-2-devnet | 227490 | 1440 |
| phase-2-testnet | 198663 | 26124 |
| phase-2-private-mainnet | 882251 | 8844 |

## Future Improvements

1. **Configuration Validation**: Add validation to ensure environment-specific values are valid.
2. **Environment-Based Defaults**: Allow setting default values for other configuration parameters based on the environment.
3. **Additional Parameters**: Add more environment-specific parameters as needed. 