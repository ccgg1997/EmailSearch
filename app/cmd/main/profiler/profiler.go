// profiler/profiler.go
package profiler

import (
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
)

var cpuFile *os.File

func StartCPUProfile() error {
	// Establecer el directorio y el nombre del archivo
	dir := "./profiler/profiles"
	filename := "cpu.pprof"
	path := filepath.Join(dir, filename)

	// Verificar si el directorio existe, si no, crearlo
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkdirErr := os.Mkdir(dir, 0755); mkdirErr != nil {
			log.Printf("Error creating directory: %v", mkdirErr)
			return mkdirErr
		}
	}

	// Crear (o sobrescribir) el archivo en el directorio especificado
	var err error
	cpuFile, err = os.Create(path)
	if err != nil {
		log.Printf("Error opening the file: %v", err)
		return err
	}

	pprof.StartCPUProfile(cpuFile)
	return nil
}

func StopCPUProfile() error {
	pprof.StopCPUProfile()
	if err := cpuFile.Close(); err != nil {
		log.Printf("Error closing file: %v", err)
		return err
	}
	return nil
}
