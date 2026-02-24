# Go SSH Menu Tool - Initial Plan

## Features
- CLI menu with two levels: Project/Customer → Server
- YAML config with global command template and per-server override
- Help menu ('h'), exit ('q')
- Add sessions interactively

## Steps
1. Design YAML config structure
2. Implement config parsing in Go
3. Build CLI menu system (promptui)
4. Command execution logic
5. Error handling and user feedback

## Verification
- Test with sample YAML configs
- Run tool and verify menu navigation and command execution
- Check error handling for invalid configs and failed SSH commands

## Decisions
- CLI interface for portability and simplicity
- YAML config for readability and ease of editing
- Global command template with per-server override for flexibility

---
This plan is the foundation for the Go SSH menu tool. Update as requirements evolve.
