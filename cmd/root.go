package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/BESTSELLER/harpocrates/validate"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/gookit/color"
	// "github.com/rs/zerolog" // zerolog is now primarily handled in config package
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	// duration will be handled by setupContinuousMode and used in the main Run loop.
	// runSuccess is a global variable to reflect the status of the fetchAndWriteSecrets operation.
	// This is used by the HTTP status endpoint in continuous mode.
	runSuccess bool
)

// ErrValidationSuccessful is a sentinel error to indicate successful validation in --validate mode.
var ErrValidationSuccessful = fmt.Errorf("validation successful")

func processInput(cmd *cobra.Command, args []string) (util.SecretJSON, error) {
	var data string
	var input util.SecretJSON

	currentSecretFile := config.GetSecretFile()
	currentSecretSlice := config.GetSecretSlice()

	if currentSecretFile != "" { // --file is being used
		fileData, err := files.Read(currentSecretFile)
		if err != nil {
			return input, fmt.Errorf("failed to read secret file '%s': %w", currentSecretFile, err)
		}
		data = fileData // Assign after successful read
		if err := validate.SecretsFile(data); err != nil {
			return input, fmt.Errorf("secret file validation failed for '%s': %w", currentSecretFile, err)
		}
		if config.Config.Validate {
			log.Info().Msgf("File validation successful for '%s'.", currentSecretFile)
			return input, ErrValidationSuccessful
		}
		parsedInput, err := util.ReadInput(data)
		if err != nil {
			return input, fmt.Errorf("failed to parse input from file '%s': %w", currentSecretFile, err)
		}
		input = parsedInput
	} else if len(currentSecretSlice) > 0 { // Parameters is being used
		if config.Config.Output == "" {
			cmd.Usage() // Show usage if output is missing for secret parameters.
			return input, fmt.Errorf("output is required when using secret parameters")
		}
		y := make([]interface{}, len(currentSecretSlice))
		for i, s := range currentSecretSlice {
			y[i] = s
		}
		input = util.SecretJSON{Secrets: y}
	} else { // inline secret is being used
		if len(args) == 0 {
			cmd.Help() // Show help if no arguments are provided for inline mode.
			return input, fmt.Errorf("no input provided: use --file, --secret, or inline JSON")
		}
		// args[0] is expected to be the inline JSON/YAML string.
		if err := validate.SecretsFile(args[0]); err != nil {
			return input, fmt.Errorf("inline secret validation failed: %w", err)
		}
		if config.Config.Validate {
			log.Info().Msg("Inline JSON validation successful.")
			return input, ErrValidationSuccessful
		}
		parsedInput, err := util.ReadInput(args[0])
		if err != nil {
			return input, fmt.Errorf("failed to parse inline input: %w", err)
		}
		input = parsedInput
	}
	// If we reach here and input is still empty (e.g. secretFile was empty string, slice was empty, args were empty)
	// it means no valid input source was found or processed.
	// The individual blocks should have returned errors if their specific input was expected but faulty.
	// If all inputs are simply not provided, the `else` block for inline secrets handles the `len(args) == 0` case.
	return input, nil
}

func setupContinuousMode() (*time.Duration, error) {
	var internalDuration time.Duration
	if config.Config.Continuous {
		intervalEnv, intervalSet := os.LookupEnv("INTERVAL")
		if !intervalSet {
			return nil, fmt.Errorf("INTERVAL environment variable must be set for continuous mode")
		}

		durationParsed, err := time.ParseDuration(intervalEnv)
		if err != nil {
			return nil, fmt.Errorf("error parsing INTERVAL: %w", err)
		}

		if durationParsed < (1 * time.Minute) {
			return nil, fmt.Errorf("interval must be at least 1 minute")
		}
		internalDuration = durationParsed
		log.Debug().Msgf("Continuous mode enabled, will run every %s", internalDuration)

		config.Config.Append = false // In continuous mode, overwrite the file
		http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			if runSuccess {
				w.WriteHeader(http.StatusOK)
			} else {
				// Consider sending http.StatusServiceUnavailable if the process is ongoing but last attempt failed
				w.WriteHeader(http.StatusTooEarly) // Or another appropriate status for "not ready yet" or "failed"
			}
		})
		go func() {
			if err := http.ListenAndServe(":8000", nil); err != nil {
				log.Error().Err(err).Msg("Failed to start status server")
			}
		}()
		return &internalDuration, nil
	}
	return nil, nil // Not in continuous mode
}

func fetchAndWriteSecrets(cmd *cobra.Command, input util.SecretJSON) (bool, error) {
	if err := vault.Login(); err != nil {
		// vault.Login now returns an error, so we handle it here.
		// Adding context to the error before passing it up.
		return false, fmt.Errorf("vault login failed: %w", err)
	}

	vaultClient := vault.NewClient() // This still uses log.Fatal internally as per subtask instructions.

	allSecrets, err := vaultClient.ExtractSecrets(input, config.Config.Append)
	if err != nil {
		return false, fmt.Errorf("failed to extract secrets: %w", err)
	}

	// This format validation should ideally be done earlier, perhaps after flag parsing.
	if cmd.Flags().Changed("format") && (config.Config.Format != "json" && config.Config.Format != "env" && config.Config.Format != "secret" && config.Config.Format != "yaml") {
		cmd.Usage()
		return false, fmt.Errorf("please use a valid format: json, env, secret, or yaml")
	}

	for _, v := range allSecrets {
		fileName := config.Config.FileName
		if v.Filename != "" {
			fileName = v.Filename
		}

		var content string
		effectiveFormat := v.Format
		if effectiveFormat == "" {
			effectiveFormat = config.Config.Format // Fallback to global format if specific one is empty
		}

		switch strings.ToLower(effectiveFormat) {
		case "json":
			jsonContent, err := v.Result.ToJSON()
			if err != nil {
				return false, fmt.Errorf("failed to format secrets to JSON for file '%s': %w", fileName, err)
			}
			content = jsonContent
		case "env":
			content = v.Result.ToENV()
		case "secret":
			content = v.Result.ToK8sSecret()
		case "yaml":
			yamlContent, err := v.Result.ToYAML()
			if err != nil {
				return false, fmt.Errorf("failed to format secrets to YAML for file '%s': %w", fileName, err)
			}
			content = yamlContent
		default:
			// It might be better to error here if the format is unknown and not empty.
			log.Warn().Msgf("Unsupported format '%s' for file '%s', defaulting to env.", effectiveFormat, fileName)
			content = v.Result.ToENV() // Defaulting for safety, but consider erroring.
		}
		if err := files.Write(config.Config.Output, fileName, content, v.Owner, config.Config.Append); err != nil {
			// Decide if one failed write should stop all, or just log and continue.
			// Returning error to stop, as this seems like a significant failure.
			return false, fmt.Errorf("failed to write secret file '%s/%s': %w", config.Config.Output, fileName, err)
		}
		log.Debug().Msgf("Secrets written to file: %s/%s", config.Config.Output, fileName)
	}
	return true, nil
}

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// processInput will call log.Fatal if it encounters an error that should stop execution.
		// This is kept to maintain existing behavior for CLI user feedback.
		// For library use, these would ideally be returned errors.
		input, err := processInput(cmd, args)
		if err != nil {
			if err == ErrValidationSuccessful {
				// This means --validate was used and validation passed.
				// Logged in processInput, so just exit normally.
				return
			}
			// For other errors from processInput, log fatal.
			log.Fatal().Err(err).Msg("Error processing input")
			return
		}

		// This explicit check for config.Config.Validate after processInput might be redundant
		// if processInput correctly returns ErrValidationSuccessful.
		// If processInput returns nil for a successful validation (as it did before ErrValidationSuccessful),
		// this block would be necessary. With ErrValidationSuccessful, the `if err == ErrValidationSuccessful`
		// block above should handle the exit for successful validation.
		// If we reach here and config.Config.Validate is true, it implies an error occurred before
		// ErrValidationSuccessful could be returned by processInput, which would be caught by log.Fatal above.
		// Thus, this specific check might only be relevant if processInput's error handling for validation changes.
		// For now, it acts as a safeguard or handles cases where processInput might return nil on successful validation
		// (which is not the current implementation with ErrValidationSuccessful).
		if config.Config.Validate {
			// This path should ideally not be hit if processInput correctly uses ErrValidationSuccessful.
			// If it is hit, it implies --validate is true but processInput didn't return ErrValidationSuccessful,
			// meaning either an actual error occurred (handled above) or validation was successful but signaled differently.
			log.Debug().Msg("Exiting after validation check.") // More generic debug message.
			return
		}

		var currentLoopDuration *time.Duration
		if config.Config.Continuous {
			var setupErr error
			currentLoopDuration, setupErr = setupContinuousMode()
			if setupErr != nil {
				log.Fatal().Err(setupErr).Msg("Error setting up continuous mode")
				return
			}
		}

		// Initialize runSuccess to false before starting the loop.
		// It will be set to true upon successful completion of fetchAndWriteSecrets.
		runSuccess = false
		for {
			// Update global runSuccess based on the outcome of fetchAndWriteSecrets
			currentIterSuccess, iterErr := fetchAndWriteSecrets(cmd, input)
			runSuccess = currentIterSuccess // Update global status for HTTP handler

			if iterErr != nil {
				log.Error().Err(iterErr).Msg("Error fetching and writing secrets")
				// In continuous mode, log the error and continue (or implement backoff).
				// For single run mode, this error will lead to exiting after the loop.
				if !config.Config.Continuous {
					// No need to fatal here, let the main logic handle exit.
					break // Exit loop on error in single run mode
				}
				// In continuous mode, we might want to sleep for the interval even on error,
				// or implement a different backoff. For now, it proceeds to sleep.
			}

			if !config.Config.Continuous {
				break // Exit loop if not in continuous mode
			}

			if currentLoopDuration != nil {
				log.Debug().Msgf("Sleeping for %s", *currentLoopDuration)
				time.Sleep(*currentLoopDuration)
			} else {
				// This case should not be reached if continuous mode is correctly set up
				log.Error().Msg("Continuous mode is active but duration is not set. Exiting loop.")
				break
			}
		}
		// After the loop, the global runSuccess reflects the status of the last attempt.
		// Log final status.
		if runSuccess {
			log.Info().Msg("Operations completed successfully.")
		} else {
			log.Error().Msg("Operations failed or completed with errors.")
			// os.Exit(1) could be used here if a non-zero exit code is critical for automation.
			// For a CLI tool, often logging the error is sufficient and letting main return naturally (or with error).
		}
	},
	Use:   "harpocrates",
	Short: fmt.Sprintf("%sThis application will fetch secrets from Hashicorp Vault", color.Blue.Sprintf("\"Harpocrates was the god of silence, secrets and confidentiality\"\n")),
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Initialize flags using the function from the config package
	config.InitFlags(rootCmd)

	// Call SyncEnvToFlags and SetupLogLevel from the config package
	// This replaces the old initConfig and direct flag definitions
	config.SyncEnvToFlags(rootCmd)
	config.SetupLogLevel()
}

// initConfig is no longer needed as its functionality has been moved to init()
// and uses functions from the config package.

// SetupLogLevel has been moved to the config package.
