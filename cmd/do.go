package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/BESTSELLER/harpocrates/config"
	"github.com/BESTSELLER/harpocrates/files"
	"github.com/BESTSELLER/harpocrates/util"
	"github.com/BESTSELLER/harpocrates/validate"
	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func loadLocalVaultToken() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Debug().Err(err).Msg("unable to get home directory")
		return
	}

	vaultTokenFile := path.Join(homeDir, ".vault-token")
	token, err := os.ReadFile(vaultTokenFile)
	if err == nil {
		config.Config.VaultToken = strings.TrimSpace(string(token))
		log.Debug().Msg("using vault token from ~/.vault-token")
	}

}

func doIt(cmd *cobra.Command, args []string, tmp bool) {
	loadLocalVaultToken()

	finalEnv := os.Environ()

	var data string
	var input util.SecretJSON

	if secretFile != "" {
		data = files.Read(secretFile)
		validFile := validate.SecretsFile(data)
		if !validFile {
			log.Fatal().Msg("Invalid file")
		}
		if config.Config.Validate {
			return
		}
		input = util.ReadInput(data)
	} else if len(*secret) > 0 {
		if config.Config.Output == "" {
			log.Error().Msg("Output is required!")
			cmd.Usage()
			return
		}

		y := make([]interface{}, len(*secret))

		for i, s := range *secret {
			y[i] = s
		}

		input = util.SecretJSON{
			Secrets: y,
		}
	} else {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		if validate.SecretsFile(args[0]) {
			input = util.ReadInput(args[0])
		}
		if config.Config.Validate {
			return
		}
	}

	if tmp {
		// Create temporary directory and file
		dir, err := os.MkdirTemp("", "harpocrates")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(dir)

		config.Config.Output = path.Join(dir, config.Config.Output)

		fmt.Println("output:", config.Config.Output)
	}

	envVar, ok := os.LookupEnv("CONTINUOUS")
	if ok && strings.ToLower(envVar) == "true" {
		config.Config.Continuous = true

		interval, _ := os.LookupEnv("INTERVAL")

		durationParsed, err := time.ParseDuration(interval)
		if err != nil {
			log.Fatal().Err(err).Msgf("%s", err)
		}

		if durationParsed < (1 * time.Minute) {
			log.Fatal().Msg("Interval must be at least 1 minute")
		}

		duration = &durationParsed
		log.Debug().Msgf("Continuous mode enabled, will run every %s", durationParsed)

		config.Config.Append = false
		http.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if success {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusTooEarly)
			}
		}))
		go func() {
			http.ListenAndServe(":8000", nil)
		}()
	}

	for {
		vault.Login()

		vaultClient := vault.NewClient()

		allSecrets, err := vaultClient.ExtractSecrets(input, config.Config.Append)
		if err != nil {
			success = false
			log.Fatal().Err(err).Msgf("%s", err)
		}

		if cmd.Flags().Changed("format") && (config.Config.Format != "json" && config.Config.Format != "env" && config.Config.Format != "secret") {
			log.Error().Msg("Please a valid format of either: json, env or secret")
			cmd.Help()
			return
		}

		for _, v := range allSecrets {
			fileName := config.Config.FileName
			if v.Filename != "" {
				fileName = v.Filename
			}

			switch v.Format {
			case "json":
				files.Write(config.Config.Output, fileName, v.Result.ToJSON(), v.Owner, config.Config.Append)
			case "env":
				files.Write(config.Config.Output, fileName, v.Result.ToENV(), v.Owner, config.Config.Append)
				finalEnv = append(finalEnv, v.Result.ToKVarray("")...)
			case "secret":
				files.Write(config.Config.Output, fileName, v.Result.ToK8sSecret(), v.Owner, config.Config.Append)
			case "yaml":
				files.Write(config.Config.Output, fileName, v.Result.ToYAML(), v.Owner, config.Config.Append)
			}
			log.Debug().Msgf("Secrets written to file: %s/%s", config.Config.Output, fileName)
		}
		success = true

		if !config.Config.Continuous {
			break
		}

		log.Debug().Msgf("Sleeping for %s", duration)
		time.Sleep(*duration)
	}

	if tmp {
		// Set up cancellable context and signal handling for ctrl+c
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-signals
			cancel()
		}()

		fmt.Println(config.Config.Output)
		// 4. Start the child application with the temporary file path using the context
		// cmd := exec.CommandContext(ctx, "bash", "-c", "echo $HEJSA")
		fmt.Println("args:", args)
		cmd := exec.CommandContext(ctx, "bash", "-c", strings.Join(args, " "))
		finalEnv = append(finalEnv, fmt.Sprintf("SECRET_PATH=%s", config.Config.Output))
		cmd.Env = finalEnv
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil || err == context.Canceled {
				// Context cancelled (e.g., ctrl+c)
				return
			}
			panic(err)
		}
	}

}
