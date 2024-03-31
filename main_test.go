package main

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	// Set up test environment
	os.Setenv("CONFIG_NAME", "test_config")
	os.Setenv("CONFIG_TYPE", "yaml")
	os.Setenv("CONFIG_PATH", ".")

	// Call the function to be tested
	initConfig()

	// Verify the expected configuration values
	expectedConfigName := "test_config"
	actualConfigName := viper.GetString("config_name")
	if actualConfigName != expectedConfigName {
		t.Errorf("Expected config_name to be %s, but got %s", expectedConfigName, actualConfigName)
	}

	expectedConfigType := "yaml"
	actualConfigType := viper.GetString("config_type")
	if actualConfigType != expectedConfigType {
		t.Errorf("Expected config_type to be %s, but got %s", expectedConfigType, actualConfigType)
	}

	expectedConfigPath := "."
	actualConfigPath := viper.GetString("config_path")
	if actualConfigPath != expectedConfigPath {
		t.Errorf("Expected config_path to be %s, but got %s", expectedConfigPath, actualConfigPath)
	}
}

func TestInitTimeZone(t *testing.T) {
	// Call the function to be tested
	initTimeZone()

	// Verify the expected time zone
	expectedTimeZone := "Asia/Bangkok"
	actualTimeZone := time.Local.String()
	if actualTimeZone != expectedTimeZone {
		t.Errorf("Expected time zone to be %s, but got %s", expectedTimeZone, actualTimeZone)
	}
}
