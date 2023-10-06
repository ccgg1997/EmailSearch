// profiler/profiler.go
package profiler

import (
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime"
)

var cpuFile *os.File
var heapFile *os.File

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


func StartHeapProfile() error {
	// Establecer el directorio y el nombre del archivo
	dir := "./profiler/profiles"
	filename := "heap.pprof"
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
	heapFile, err = os.Create(path)
	if err != nil {
		log.Printf("Error opening the file: %v", err)
		return err
	}

	return nil
}

func StopHeapProfile() error {
	// Llama a GC para obtener estadísticas de memoria más precisas
	runtime.GC()

	// Escribe el perfil de heap en el archivo
	if err := pprof.WriteHeapProfile(heapFile); err != nil {
		log.Printf("Error writing heap profile: %v", err)
		return err
	}

	// Cierra el archivo
	if err := heapFile.Close(); err != nil {
		log.Printf("Error closing file: %v", err)
		return err
	}

	return nil
}
