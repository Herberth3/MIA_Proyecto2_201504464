package comandos

import (
	strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

type PARTICIONMONTADA struct {
	Letra  [1]byte
	Estado int
	Nombre string
}

type DISCOMONTADO struct {
	Path        string
	Numero      int
	Estado      int
	Particiones [99]PARTICIONMONTADA
}

// Arreglo que contendra informacion sobre los discos y sus particiones montadas
var Discos [26]DISCOMONTADO

// Solo 26 letras y 99 posiciones, pues se espera que al EVALUARLO no se monten muchas particiones
var Letras = [99]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

func EjecutarMOUNT(path string, name string) {

	// Se limpia la cadena Consola para recolectar informacion de una nueva ejecucion
	Consola = ""
	part_startExtendida := 0
	existePath := false
	disk_mount_pos := 0

	// Si el disco es abierto con exito, se empieza a verificar si se puede montar la particion
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	mbr_disco := strct.MBR{}
	is_Partition := false
	index_Primaria := 0

	// --------Se extrae el MBR del disco---------
	var size int = int(binary.Size(mbr_disco))
	disco_actual.Seek(0, 0)
	data := leerEnFILE(disco_actual, size)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &mbr_disco)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	for i := 0; i < 4; i++ {

		if string(mbr_disco.Mbr_partition[i].Part_type[:]) == "E" {
			part_startExtendida = byteToInt(mbr_disco.Mbr_partition[i].Part_start[:])
		}

		if byteToStr(mbr_disco.Mbr_partition[i].Part_name[:]) == name {
			is_Partition = true
			index_Primaria = i
		}
	}

	// Se identifica el part_status con un '2' de la particion primaria reconocida para no interactuar con ella
	// En casos como en ADD no apareceria por cuestion de modificacion alguna estando esta montada
	if is_Partition {
		copy(mbr_disco.Mbr_partition[index_Primaria].Part_status[:], "2")
		disco_actual.Seek(0, 0)
		// --------------- Se guarda el mbr actualizado -------------
		s1 := &mbr_disco
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, s1)
		escribirDentroFILE(disco_actual, binario3.Bytes())
	}

	// Verificar si es una particion logica
	// Se busca dentro de la particion Extendida
	if !(is_Partition) && part_startExtendida != 0 {
		ebr_temporal := strct.EBR{}

		// Se posiciona en el partstart del primer EBR en al Extendida
		disco_actual.Seek(int64(part_startExtendida), 0)
		// --------Se extrae el EBR del disco---------
		var ebrsize int = int(binary.Size(ebr_temporal))
		data := leerEnFILE(disco_actual, ebrsize)
		buffer := bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.BigEndian, &ebr_temporal)
		if err != nil {
			Consola += "Binary.Read failed\n"
			msg_error(err)
		}

		for string(ebr_temporal.Part_status[:]) != "1" && byteToInt(ebr_temporal.Part_next[:]) != -1 {

			if strings.Compare(byteToStr(ebr_temporal.Part_name[:]), name) == 0 {
				is_Partition = true
			}

			var ebrPartNext int = byteToInt(ebr_temporal.Part_next[:])
			disco_actual.Seek(int64(ebrPartNext), 0)

			data := leerEnFILE(disco_actual, ebrsize)
			buffer := bytes.NewBuffer(data)
			err = binary.Read(buffer, binary.BigEndian, &ebr_temporal)
			if err != nil {
				Consola += "Binary.Read failed\n"
				msg_error(err)
			}
		}
	}

	if !(is_Partition) {
		Consola += "Error. La particion no existe.\n"
		disco_actual.Close()
		return
	}

	disco_actual.Close()

	// YA SE HA VERIFICADO QUE LA PARTICION SI EXISTE, AHORA SE VERIFICA SI YA ESTA MONTADA O SE PUEDE MONTAR

	// SE VERIFICA SI LA PARTICION YA HA SIDO MONTADA CON ANTERIORIDAD
	for i := 0; i < 26; i++ {

		// El path valida si el disco esta montado, pues se ha guardado en la estructura
		// Se verifica si ya existe el disco montado (path)
		if strings.Compare(Discos[i].Path, path) == 0 {

			for j := 0; j < 99; j++ {

				if strings.Compare(Discos[i].Particiones[j].Nombre, name) == 0 && Discos[i].Particiones[j].Estado == 1 {
					Consola += "Error: Particion ya montada: 64" + strconv.Itoa(Discos[i].Numero) + string(Discos[i].Particiones[j].Letra[:]) + "\n"
					return
				}
			}

			existePath = true
			disk_mount_pos = i
		}
	}

	// LA PARTICION AUN NO HA SIDO MONTADA

	// Si no existe el disco montado, se monta el disco; aun no se monta la particion
	if !existePath {

		for i := 0; i < 26; i++ {

			// Al siguiente disco con estado '0' se llena de informacion y se activa el montaje del path
			if Discos[i].Estado == 0 {
				// Se actualizan los atributos del struct DISCOMONTADO
				Discos[i].Estado = 1
				// NUMERO que identifica al disco al que pertenecen las particiones
				Discos[i].Numero = i + 1
				//copy(Discos[i].Path[:], path)
				Discos[i].Path = path
				// Posicion del disco montado
				disk_mount_pos = i
				existePath = true
				break

			}
		}
	}

	// Disco ya montado, solo se agrega la particion
	if existePath {

		for i := 0; i < 99; i++ {

			if Discos[disk_mount_pos].Particiones[i].Estado == 0 {

				Discos[disk_mount_pos].Particiones[i].Estado = 1
				// Se le asigna una LETRA la letra de particion, para identificar las particiones montadas
				copy(Discos[disk_mount_pos].Particiones[i].Letra[:], Letras[i])
				//copy(Discos[disk_mount_pos].Particiones[i].Nombre[:], name)
				Discos[disk_mount_pos].Particiones[i].Nombre = name

				Consola += "Particion montada exitosamente, id: 64" + strconv.Itoa(Discos[disk_mount_pos].Numero) + string(Discos[disk_mount_pos].Particiones[i].Letra[:]) + "\n"
				break
			}
		}
	}

	show_montajes()
}

func show_montajes() {
	Consola += "<<<-------------------------- MONTAJES ---------------------->>>\n"
	for i := 0; i < 26; i++ {

		for j := 0; j < 99; j++ {

			if Discos[i].Particiones[j].Estado == 1 {
				Consola += "64" + strconv.Itoa(Discos[i].Numero) + string(Discos[i].Particiones[j].Letra[:]) + "\n"

			}
		}
	}
}
