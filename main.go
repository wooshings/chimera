package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	grab "github.com/cavaliergopher/grab/v3"
)

type ModList struct {
	Both    []string `json:"both"`
	Client  []string `json:"client"`
	Server  []string `json:"server"`
	Version string   `json:"version"`
	Loader  string   `json:"loader"`
}

type ModData struct {
	Name         string           `json:"name"`
	ID           string           `json:"project_id"`
	Files        []FileData       `json:"files"`
	Versions     []string         `json:"game_versions"`
	Loaders      []string         `json:"loaders"`
	Dependencies []DependencyData `json:"dependencies"`
}

type DependencyData struct {
	VersionId string `json:"version_id"`
	ProjectId string `json:"project_id"`
	Filename  string `json:"file_name"`
	Type      string `json:"dependency_type"`
}

type FileData struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
}

type Config struct {
	ClientType string   `json:"client_type"`
	ModList    []string `json:"mod_lists"`
}

var args []string = os.Args[1:]

func main() {
	var config Config
	fileContent, _ := os.ReadFile("chimera_conf.json")
	json.Unmarshal(fileContent, &config)

	for _, v := range config.ModList {
		var mod_list ModList = get_mod_list(v)

		mod_list_for_type(config.ClientType, mod_list)
		mod_list_for_type("both", mod_list)

	}

	fmt.Println("\nDone.")
}

func mod_list_for_type(client_type string, mod_list ModList) {
	var list []string
	switch client_type {
	case "client":
		list = mod_list.Client
	case "server":
		list = mod_list.Server
	case "both":
		list = mod_list.Both
	}
	fmt.Println(list)

	for _, v := range list {
		fmt.Println(v)
	}

	fmt.Println("")
	for _, v := range list {
		var mod_data []ModData = search_mod(v, mod_list.Loader, mod_list.Version)
		if len(mod_data) == 0 {
			fmt.Println("No valid version found for " + v + ". Skipping...")
			continue
		}
		var mod = check_for_updates(mod_data)

		download_mod(mod, mod_list)

	}
}

func get_mod_list(mod_list_url string) ModList {
	fmt.Println(mod_list_url)
	if isValidUrl(mod_list_url) {
		var mod_list ModList
		mod_list, _ = json_from_request[ModList](mod_list_url)

		return mod_list
	} else {
		var mod_list ModList
		var fileContent []byte
		fileContent, _ = os.ReadFile(mod_list_url)
		json.Unmarshal(fileContent, &mod_list)

		return mod_list
	}
}

var search_times int

func search_mod(slug string, loader string, version string) []ModData {
	fmt.Println("Searching for", slug, loader, version)
	if search_times == 1 {
		return []ModData{}
	}
	slug = strings.ReplaceAll(slug, " ", "-")
	var furl = fmt.Sprint("https://api.modrinth.com/v2/project/", slug, "/version")
	var url, _ = url.Parse(furl)
	var params = url.Query()
	params.Add("loaders", `["`+loader+`"]`)
	params.Add("game_versions", `["`+version+`"]`)
	url.RawQuery = params.Encode()
	var mod_data, _ = json_from_request[[]ModData](url.String())
	if len(mod_data) == 0 {
		search_times += 1
		slug = strings.ReplaceAll(slug, " ", "")
		search_mod(slug, loader, version)
	}
	search_times = 0
	return mod_data
}

// "https://api.modrinth.com/v2/project/{slug}/version?loaders=[\"{loader}\"]&game_versions=[\"{version}\"]",
// https://api.modrinth.com/v2/project/fabric-api/version?loaders=["forge"]&game_versions=["version"]
// "https://api.modrinth.com/v2/project/fabric-api/version

func check_for_updates(mod_data []ModData) ModData {
	var entries, _ = os.ReadDir("mods")

	for _, entry := range entries {
		for i, mod := range mod_data {
			//fmt.Println(mod.Files[0].Filename)
			if entry.Name() == mod.Files[0].Filename {
				if i != 0 {
					fmt.Println(mod.Name, "is not most recent version. Updating...")
					os.Remove(fmt.Sprint("mods/", mod.Files[0].Filename))
				}
				continue
			}
		}
	}

	return mod_data[0]
}

func download_mod(mod_data ModData, mod_list ModList) {
	var entries, _ = os.ReadDir("mods")
	var file_found bool = false
	for _, entry := range entries {
		if entry.Name() == mod_data.Files[0].Filename {
			file_found = true
		}
	}

	if file_found {
		return
	}

	fmt.Println("\nDownloading", mod_data.Name, "...")

	_, err := grab.Get("mods/", mod_data.Files[0].Url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File saved as", mod_data.Files[0].Filename)

	if len(mod_data.Dependencies) > 0 {
		fmt.Println("Dependencies found. Installing...")
		for _, v := range mod_data.Dependencies {
			fmt.Println(v.ProjectId, "is", v.Type, "for", mod_data.Name)
			if v.Type != "required" {
				continue
			}
			var mod_data []ModData = search_mod(v.ProjectId, mod_list.Loader, mod_list.Version)
			if len(mod_data) == 0 {
				fmt.Println("No valid version found for " + v.Filename + ". Skipping...")
				continue
			}
			var mod = check_for_updates(mod_data)

			download_mod(mod, mod_list)
		}
	}
}

func clean(mod_data ModData) {
	var entries, _ = os.ReadDir("mods")
	for _, entry := range entries {
		if entry.Name() == mod_data.Files[0].Filename {
			os.Remove(fmt.Sprint("mods/", entry.Name()))
			continue
		}
	}
}
