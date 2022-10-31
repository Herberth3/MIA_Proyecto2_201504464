package comandos

import (
	strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Variable tipo string que almacenara mensajes de error o exito al ejecutar el comando
var Consola string

func msg_error(err error) {
	Consola += "Error: " + err.Error() + "\n"
}

func EjecutarMKDISK(size int, unit string, fit string, path string) {

	// Se limpia la cadena Consola para recolectar informacion de una nueva ejecucion
	Consola = ""
	tamanoDisk := 1024
	ajuste := "F"
	bloque := make([]byte, 1024)
	limite := 0
	// Instancia de la estructura MBR
	mbr := strct.MBR{}

	// Verificamos si el disco ya existe
	validar, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err == nil {
		Consola += "Disco ya existente, intente con un nombre diferente\n"
		validar.Close()
		return
	}

	// Si el disco y la ruta no existen se intenta crear el directorio pero aun no el archivo del disco
	crearDirectorio(path)

	// Estableciendo la fecha de la creacion del disco
	t := time.Now()
	fechayhora := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	// Estableciendo signature (identificador unico)  del disco
	randomInteger := rand.Intn(200)

	// Estableciendo el tamaño del disco
	// NOTA: A la variable tamanoDisk aun le falta multiplicarse por 1024 para cumplir los bytes establecidos
	if unit == "m" {
		// La unidad del tamaño esta en Megabytes
		tamanoDisk = size * 1024
	} else {
		// La unida del tamaño esta en Kilobytes
		tamanoDisk = size
	}

	// Estableciendo el ajuste del disco en el MBR
	if fit == "bf" {
		ajuste = "B"
	} else if fit == "ff" {
		ajuste = "F"
	} else {
		ajuste = "W"
	}

	// Preparacion del bloque a escribir en archivo
	for i := 0; i < 1024; i++ {
		bloque[i] = 0
	}

	// Crear disco fisico
	Consola += "Creando disco, espere.......\n"
	// Creacion, escritura y cierre de archivo
	disco, err := os.Create(path)
	if err != nil {
		msg_error(err)
		return
	}

	for limite < tamanoDisk {
		_, err := disco.Write(bloque)
		if err != nil {
			msg_error(err)
		}
		limite++
	}

	disco.Close()

	// Se lee el archivo de manera binaria para escribir dentro el MBR creado a continuacion
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
	}

	// Escritura en el archivo utilizando SEEK_START

	// Asignando valores a la estructura MBR que se guardara en el archivo binario (disco)
	copy(mbr.Mbr_size[:], strconv.Itoa(tamanoDisk*1024))
	copy(mbr.Mbr_date_created[:], fechayhora)
	copy(mbr.Mbr_disk_signature[:], strconv.Itoa(randomInteger))
	copy(mbr.Mbr_disk_fit[:], ajuste)

	// Iniciando con ceros los atributos de cada particion
	// En part_status de la particion: '0' = sin mkfs, '1' = con mkfs
	// El atributo mbr_partition del MBR es un arreglo de 4 posiciones
	for i := 0; i < 4; i++ {
		copy(mbr.Mbr_partition[i].Part_status[:], "0")
		copy(mbr.Mbr_partition[i].Part_type[:], "0")
		copy(mbr.Mbr_partition[i].Part_fit[:], "0")
		copy(mbr.Mbr_partition[i].Part_size[:], strconv.Itoa(0))
		copy(mbr.Mbr_partition[i].Part_start[:], strconv.Itoa(-1))
		copy(mbr.Mbr_partition[i].Part_name[:], "")
	}

	disco_actual.Seek(0, 0) // nos posicionamos en el inicio del archivo.
	s1 := &mbr
	//Escribimos struct.
	var binario3 bytes.Buffer
	binary.Write(&binario3, binary.BigEndian, s1)
	escribirDentroFILE(disco_actual, binario3.Bytes())

	disco_actual.Close()

	Consola += "Disco creado con exito\n"

}

func crearDirectorio(ruta string) {

	last_index_barra := strings.LastIndex(ruta, "/")
	new_Directorio := ruta[:last_index_barra]

	if _, err := os.Stat(new_Directorio); os.IsNotExist(err) {
		err = os.Mkdir(new_Directorio, 0755)
		if err != nil {
			// Aquí puedes manejar mejor el error, es un ejemplo
			msg_error(err)
		}
	}
}
