package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"dops/internal/adapters"
	"dops/internal/config"
	"dops/internal/crypto"

	"github.com/spf13/cobra"
)

func newConfigCmd(dopsDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write dops configuration",
	}

	configPath := filepath.Join(dopsDir, "config.json")
	keysDir := filepath.Join(dopsDir, "keys")
	fs := adapters.NewOSFileSystem()
	store := config.NewFileStore(fs, configPath)

	cmd.AddCommand(newConfigSetCmd(store, keysDir))
	cmd.AddCommand(newConfigGetCmd(store))
	cmd.AddCommand(newConfigUnsetCmd(store))
	cmd.AddCommand(newConfigListCmd(store, keysDir))

	return cmd
}

func newConfigSetCmd(store *config.FileConfigStore, keysDir string) *cobra.Command {
	var secret bool

	cmd := &cobra.Command{
		Use:   "set key=value",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.SplitN(args[0], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("expected key=value, got %q", args[0])
			}
			key, value := parts[0], parts[1]

			cfg, err := store.Load()
			if err != nil {
				cfg, err = store.EnsureDefaults()
				if err != nil {
					return err
				}
			}

			var finalValue any = value
			if secret {
				enc, err := crypto.NewAgeEncrypter(keysDir)
				if err != nil {
					return fmt.Errorf("init encryption: %w", err)
				}
				encrypted, err := enc.Encrypt(value)
				if err != nil {
					return fmt.Errorf("encrypt: %w", err)
				}
				finalValue = encrypted
			}

			if err := config.Set(cfg, key, finalValue); err != nil {
				return err
			}

			return store.Save(cfg)
		},
	}

	cmd.Flags().BoolVar(&secret, "secret", false, "Encrypt the value before storing")
	return cmd
}

func newConfigGetCmd(store *config.FileConfigStore) *cobra.Command {
	return &cobra.Command{
		Use:   "get key",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := store.Load()
			if err != nil {
				return err
			}

			val, err := config.Get(cfg, args[0])
			if err != nil {
				return err
			}

			if s, ok := val.(string); ok && crypto.IsEncrypted(s) {
				fmt.Fprintln(cmd.OutOrStdout(), "****")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		},
	}
}

func newConfigUnsetCmd(store *config.FileConfigStore) *cobra.Command {
	return &cobra.Command{
		Use:   "unset key",
		Short: "Remove a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := store.Load()
			if err != nil {
				return err
			}

			if err := config.Unset(cfg, args[0]); err != nil {
				return err
			}

			return store.Save(cfg)
		},
	}
}

func newConfigListCmd(store *config.FileConfigStore, keysDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Display the full configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := store.Load()
			if err != nil {
				return err
			}

			masked := crypto.MaskSecrets(cfg)
			data, err := json.MarshalIndent(masked, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}
