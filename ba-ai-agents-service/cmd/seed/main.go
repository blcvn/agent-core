package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	agentPath := os.Getenv("AGENT_CONFIG_PATH")
	if agentPath == "" {
		agentPath = "config/agents"
	}
	skillPath := os.Getenv("SKILL_CONFIG_PATH")
	if skillPath == "" {
		skillPath = "config/skills"
	}

	log.Printf("Seeding agents from %s...", agentPath)
	if err := seedAgents(agentPath); err != nil {
		log.Fatalf("Failed to seed agents: %v", err)
	}

	log.Printf("Seeding skills from %s...", skillPath)
	if err := seedSkills(skillPath); err != nil {
		log.Fatalf("Failed to seed skills: %v", err)
	}

	log.Println("Seeding complete.")
}

func seedAgents(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" {
			log.Printf("Found agent config: %s", path)
			// Parse and save logic here
			var data map[string]interface{}
			content, _ := os.ReadFile(path)
			yaml.Unmarshal(content, &data)
			log.Printf("Parsed agent: %v", data["name"])
		}
		return nil
	})
}

func seedSkills(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".yaml" {
			log.Printf("Found skill config: %s", path)
			// Parse logic
		}
		return nil
	})
}
