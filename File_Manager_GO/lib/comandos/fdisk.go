package comandos

import (
	strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func EjecutarFDISK(size int, unit string, path string, typeP string, fit string, name string) {

	// Se limpia la cadena Consola para recolectar informacion de una nueva ejecucion
	Consola = ""

	if typeP == "p" {
		crearParticionPrimaria(size, unit, path, fit, name)
	} else if typeP == "e" {
		crearParticionExtendida(size, unit, path, fit, name)
	} else {
		crearParticionLogica(size, unit, path, fit, name)
	}

	show_Particiones(path)
}

func crearParticionPrimaria(tamano int, unidad string, path string, ajuste string, nombreP string) {
	auxFit := strings.ToUpper(ajuste[:1])
	size_bytes := 1024
	buf := "1"
	masterboot := strct.MBR{}

	if unidad == "m" {
		size_bytes = tamano * 1048576
	} else if unidad == "k" {
		size_bytes = tamano * 1024
	} else {
		// Tamaño en bytes
		size_bytes = tamano
	}

	// Si el disco es abierto con exito, se empieza a verificar si se puede crear la particion
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	flagPartition := false //Flag para ver si hay una nueva particion disponibl
	numPartition := 0      // Que numero de particion es

	// Obtenemos el size del mbr
	var size int = int(binary.Size(masterboot))
	// Leemos la cantidad de bytes en el disco
	disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
	data := leerEnFILE(disco_actual, size)
	// Convierte la data en un buffer,necesario para
	// Decodificar binario
	buffer := bytes.NewBuffer(data)
	// Decodificamos y guardamos en la instancia del MBR
	err = binary.Read(buffer, binary.BigEndian, &masterboot)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// Verificar si existe una particion disponible
	// Esta puede no estar ocupada o estar inactiva/vacia y que su tamaño sea el suficiente para almacenar la nueva particion
	for i := 0; i < 4; i++ {
		if byteToInt(masterboot.Mbr_partition[i].Part_start[:]) == -1 || (string(masterboot.Mbr_partition[i].Part_status[:]) == "1" && byteToInt(masterboot.Mbr_partition[i].Part_size[:]) >= size_bytes) {
			flagPartition = true
			numPartition = i
			break
		}
	}

	// Si existe una particion disponible
	if flagPartition {
		// Verificar el espacio libre del disco
		espacioUsado := 0
		for i := 0; i < 4; i++ {
			if string(masterboot.Mbr_partition[i].Part_status[:]) != "1" {
				espacioUsado += byteToInt(masterboot.Mbr_partition[i].Part_size[:])
			}
		}

		// Reporte del espacio necesario para crear la particion
		resto := fmt.Sprintf("%d", byteToInt(masterboot.Mbr_size[:])-espacioUsado)
		Consola += "Espacio disponible: " + resto + " Bytes\n"
		Consola += "Espacio necesario: " + strconv.Itoa(size_bytes) + " Bytes\n"

		// Verificar que haya espacio suficiente para crear la particion
		if (byteToInt(masterboot.Mbr_size[:]) - espacioUsado) >= size_bytes {

			// Si, se puede crear la particion
			// Verificar en las P y en las L de la E, si existe el nombre
			if !(existePartition(masterboot, nombreP)) {

				// Se crea la particion implementando el primer ajuste
				if string(masterboot.Mbr_disk_fit[:]) == "F" {

					copy(masterboot.Mbr_partition[numPartition].Part_type[:], "P")
					copy(masterboot.Mbr_partition[numPartition].Part_fit[:], auxFit)
					// La primera posicion esta disponible
					if numPartition == 0 {
						// Part_start = se posiciona al final del tamaño del MBR
						var sizeMBR int = int(binary.Size(masterboot))
						copy(masterboot.Mbr_partition[numPartition].Part_start[:], strconv.Itoa(sizeMBR))

					} else {
						// Se crea en la primera posicion (diferente a la primera) que haya encontrado disponible
						// part_start = se posiciona al final del tamaño de la particion anterior
						var part_start int = byteToInt(masterboot.Mbr_partition[numPartition-1].Part_start[:]) + byteToInt(masterboot.Mbr_partition[numPartition-1].Part_size[:])
						copy(masterboot.Mbr_partition[numPartition].Part_start[:], strconv.Itoa(part_start))
					}

					copy(masterboot.Mbr_partition[numPartition].Part_size[:], strconv.Itoa(size_bytes))
					copy(masterboot.Mbr_partition[numPartition].Part_status[:], "0")
					copy(masterboot.Mbr_partition[numPartition].Part_name[:], nombreP)

					disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
					s1 := &masterboot
					// Escribimos struct.
					var binario3 bytes.Buffer
					binary.Write(&binario3, binary.BigEndian, s1)
					escribirDentroFILE(disco_actual, binario3.Bytes())

					// Se posiciona en el part_start de la nueva particion en el disco
					// Marca con un 1 en el archivo, indicando donde comienza la particion
					var ppart_start int = byteToInt(masterboot.Mbr_partition[numPartition].Part_start[:])
					for k := 0; k < 1; k++ {
						// Cambio de posicion de puntero dentro del archivo
						newpos, err := disco_actual.Seek(int64(k+ppart_start), 0)
						if err != nil {
							msg_error(err)
						}
						// Escritura de struct en archivo binario
						_, err = disco_actual.WriteAt([]byte(buf), newpos)
						if err != nil {
							msg_error(err)
						}
					}
					Consola += "Particion primaria creada con exito\n"
					// Se crea la particion implementando el mejor ajuste
				} else if string(masterboot.Mbr_disk_fit[:]) == "B" {
					// Se busca la mejor posicion para almacenar la particion
					bestIndex := numPartition
					for i := 0; i < 4; i++ {
						if byteToInt(masterboot.Mbr_partition[i].Part_start[:]) == -1 || (string(masterboot.Mbr_partition[i].Part_status[:]) == "1" && byteToInt(masterboot.Mbr_partition[i].Part_size[:]) >= size_bytes) {
							if i != numPartition {
								if byteToInt(masterboot.Mbr_partition[bestIndex].Part_size[:]) > byteToInt(masterboot.Mbr_partition[i].Part_size[:]) {
									bestIndex = i
									break
								}
							}
						}
					}

					copy(masterboot.Mbr_partition[bestIndex].Part_type[:], "P")
					copy(masterboot.Mbr_partition[bestIndex].Part_fit[:], auxFit)

					if bestIndex == 0 {
						var sizeMBR int = int(binary.Size(masterboot))
						copy(masterboot.Mbr_partition[bestIndex].Part_start[:], strconv.Itoa(sizeMBR))
					} else {
						var part_start int = byteToInt(masterboot.Mbr_partition[bestIndex-1].Part_start[:]) + byteToInt(masterboot.Mbr_partition[bestIndex-1].Part_size[:])
						copy(masterboot.Mbr_partition[bestIndex].Part_start[:], strconv.Itoa(part_start))
					}

					copy(masterboot.Mbr_partition[bestIndex].Part_size[:], strconv.Itoa(size_bytes))
					copy(masterboot.Mbr_partition[bestIndex].Part_status[:], "0")
					copy(masterboot.Mbr_partition[bestIndex].Part_name[:], nombreP)
					// Se guarda el MBR actualizado
					disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
					s1 := &masterboot
					// Escribimos struct.
					var binario3 bytes.Buffer
					binary.Write(&binario3, binary.BigEndian, s1)
					escribirDentroFILE(disco_actual, binario3.Bytes())

					// Se posiciona en el part_start de la nueva particion en el disco
					// Marca con un 1 en el archivo, indicando donde comienza la particion
					var ppart_start int = byteToInt(masterboot.Mbr_partition[bestIndex].Part_start[:])
					for k := 0; k < 1; k++ {
						// Cambio de posicion de puntero dentro del archivo
						newpos, err := disco_actual.Seek(int64(k+ppart_start), 0)
						if err != nil {
							msg_error(err)
						}
						// Escritura de struct en archivo binario
						_, err = disco_actual.WriteAt([]byte(buf), newpos)
						if err != nil {
							msg_error(err)
						}
					}
					Consola += "Particion primaria creada con exito\n"

					// Se crea la particion implementando el peor ajuste
				} else if string(masterboot.Mbr_disk_fit[:]) == "W" {
					// Se busca el peor ajuste
					worstIndex := numPartition
					for i := 0; i < 4; i++ {
						if byteToInt(masterboot.Mbr_partition[i].Part_start[:]) == -1 || (string(masterboot.Mbr_partition[i].Part_status[:]) == "1" && byteToInt(masterboot.Mbr_partition[i].Part_size[:]) >= size_bytes) {
							if i != numPartition {
								if byteToInt(masterboot.Mbr_partition[worstIndex].Part_size[:]) < byteToInt(masterboot.Mbr_partition[i].Part_size[:]) {
									worstIndex = i
									break
								}
							}
						}
					}

					copy(masterboot.Mbr_partition[worstIndex].Part_type[:], "P")
					copy(masterboot.Mbr_partition[worstIndex].Part_fit[:], auxFit)

					if worstIndex == 0 {
						var sizeMBR int = int(binary.Size(masterboot))
						copy(masterboot.Mbr_partition[worstIndex].Part_start[:], strconv.Itoa(sizeMBR))
					} else {
						var part_start int = byteToInt(masterboot.Mbr_partition[worstIndex-1].Part_start[:]) + byteToInt(masterboot.Mbr_partition[worstIndex-1].Part_size[:])
						copy(masterboot.Mbr_partition[worstIndex].Part_start[:], strconv.Itoa(part_start))
					}

					copy(masterboot.Mbr_partition[worstIndex].Part_size[:], strconv.Itoa(size_bytes))
					copy(masterboot.Mbr_partition[worstIndex].Part_status[:], "0")
					copy(masterboot.Mbr_partition[worstIndex].Part_name[:], nombreP)
					// Se guarda el MBR actualizado
					disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
					s1 := &masterboot
					// Escribimos struct.
					var binario3 bytes.Buffer
					binary.Write(&binario3, binary.BigEndian, s1)
					escribirDentroFILE(disco_actual, binario3.Bytes())

					// Se posiciona en el part_start de la nueva particion en el disco
					// Marca con un 1 en el archivo, indicando donde comienza la particion
					var ppart_start int = byteToInt(masterboot.Mbr_partition[worstIndex].Part_start[:])
					for k := 0; k < 1; k++ {
						// Cambio de posicion de puntero dentro del archivo
						newpos, err := disco_actual.Seek(int64(k+ppart_start), 0)
						if err != nil {
							msg_error(err)
						}
						// Escritura de struct en archivo binario
						_, err = disco_actual.WriteAt([]byte(buf), newpos)
						if err != nil {
							msg_error(err)
						}
					}
					Consola += "Particion primaria creada con exito\n"
				}
			} else {
				Consola += "ERROR: Ya existe una particion con ese nombre\n"
			}
		} else {
			Consola += "ERROR: La particion a crear excede el espacio libre\n"
		}

	} else {
		Consola += "ERROR: Ya existen 4 particiones, no se puede crear otra\n"
		Consola += "Elimine alguna para poder crear\n"
	}
	disco_actual.Close()
}

func crearParticionExtendida(tamano int, unidad string, path string, ajuste string, nombreP string) {
	auxFit := strings.ToUpper(ajuste[:1])
	size_bytes := 1024
	buf := "1"
	masterboot := strct.MBR{}

	if unidad == "m" {
		size_bytes = tamano * 1048576
	} else if unidad == "k" {
		size_bytes = tamano * 1024
	} else {
		// Tamaño en bytes
		size_bytes = tamano
	}

	// Si el disco es abierto con exito, se empieza a verificar si se puede crear la particion
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	flagPartition := false // Flag para ver si hay una nueva particion disponible
	flagExtendida := false // Flag para ver si ya hay una particion extendida
	numPartition := 0      // Que numero de particion es

	// --------Se extrae el MBR del disco---------
	// Obtenemos el size del mbr
	var size int = int(binary.Size(masterboot))
	// Leemos la cantidad de bytes en el disco
	disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
	data := leerEnFILE(disco_actual, size)
	// Convierte la data en un buffer,necesario para
	// Decodificar binario
	buffer := bytes.NewBuffer(data)
	// Decodificamos y guardamos en la instancia del MBR
	err = binary.Read(buffer, binary.BigEndian, &masterboot)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// Con el MBR extraido, se busca si existe la particion extendida
	for i := 0; i < 4; i++ {
		if string(masterboot.Mbr_partition[i].Part_type[:]) == "E" {
			flagExtendida = true
			break
		}
	}

	// Si no existe una particion extendida
	if !flagExtendida {
		// Verificar si existe una particion disponble
		for i := 0; i < 4; i++ {
			if byteToInt(masterboot.Mbr_partition[i].Part_start[:]) == -1 || (string(masterboot.Mbr_partition[i].Part_status[:]) == "1" && byteToInt(masterboot.Mbr_partition[i].Part_size[:]) >= size_bytes) {
				flagPartition = true
				numPartition = i
				break
			}
		}

		if flagPartition {
			// Verificar el espacio libre del disco
			espacioUsado := 0
			for i := 0; i < 4; i++ {
				if string(masterboot.Mbr_partition[i].Part_status[:]) != "1" {
					espacioUsado += byteToInt(masterboot.Mbr_partition[i].Part_size[:])
				}
			}

			// Reporte del espacio necesario para crear la particion
			resto := fmt.Sprintf("%d", byteToInt(masterboot.Mbr_size[:])-espacioUsado)
			Consola += "Espacio disponible: " + resto + " Bytes\n"
			Consola += "Espacio necesario: " + strconv.Itoa(size_bytes) + " Bytes\n"

			// Verificar que haya espacio suficiente para crear la particion
			if (byteToInt(masterboot.Mbr_size[:]) - espacioUsado) >= size_bytes {

				// Si, se puede crear la particion
				// Verificar en las P y en las L de la E, si existe el nombre
				if !(existePartition(masterboot, nombreP)) {

					// Se crea la particion implementando el primer ajuste
					if string(masterboot.Mbr_disk_fit[:]) == "F" {

						copy(masterboot.Mbr_partition[numPartition].Part_type[:], "E")
						copy(masterboot.Mbr_partition[numPartition].Part_fit[:], auxFit)
						// La primera posicion esta disponible
						if numPartition == 0 {
							// Part_start = se posiciona al final del tamaño del MBR
							var sizeMBR int = int(binary.Size(masterboot))
							copy(masterboot.Mbr_partition[numPartition].Part_start[:], strconv.Itoa(sizeMBR))

						} else {
							// Se crea en la primera posicion (diferente a la primera) que haya encontrado disponible
							// part_start = se posiciona al final del tamaño de la particion anterior
							var part_start int = byteToInt(masterboot.Mbr_partition[numPartition-1].Part_start[:]) + byteToInt(masterboot.Mbr_partition[numPartition-1].Part_size[:])
							copy(masterboot.Mbr_partition[numPartition].Part_start[:], strconv.Itoa(part_start))
						}

						copy(masterboot.Mbr_partition[numPartition].Part_size[:], strconv.Itoa(size_bytes))
						copy(masterboot.Mbr_partition[numPartition].Part_status[:], "0")
						copy(masterboot.Mbr_partition[numPartition].Part_name[:], nombreP)

						// Se guarda el MBR actualizado
						disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
						s1 := &masterboot
						// Escribimos struct.
						var binario3 bytes.Buffer
						binary.Write(&binario3, binary.BigEndian, s1)
						escribirDentroFILE(disco_actual, binario3.Bytes())

						// Se posiciona en el part_start de la nueva particion en el disco
						var ppart_start int = byteToInt(masterboot.Mbr_partition[numPartition].Part_start[:])
						disco_actual.Seek(int64(ppart_start), 0)
						// Se guarda la particion extendida
						extendedBoot := strct.EBR{}
						copy(extendedBoot.Part_fit[:], auxFit)
						copy(extendedBoot.Part_status[:], "0")
						copy(extendedBoot.Part_start[:], strconv.Itoa(byteToInt(masterboot.Mbr_partition[numPartition].Part_start[:])))
						copy(extendedBoot.Part_size[:], strconv.Itoa(0))
						copy(extendedBoot.Part_next[:], strconv.Itoa(-1))
						copy(extendedBoot.Part_name[:], "")

						// Se escribe el EBR en el disco, en el part_start de la particion ubicado ya con fseek
						s2 := &extendedBoot
						// Escribimos struct.
						var binario4 bytes.Buffer
						binary.Write(&binario4, binary.BigEndian, s2)
						escribirDentroFILE(disco_actual, binario4.Bytes())

						// Se escribe 1 despues del EBR, indicando donde comienza la particion
						//var sizeEBR int = int(binary.Size(extendedBoot))
						for k := 0; k < 1; k++ {
							// Cambio de posicion de puntero dentro del archivo
							newpos, err := disco_actual.Seek(int64(k+ppart_start), 0)
							if err != nil {
								msg_error(err)
							}
							// Escritura de struct en archivo binario
							_, err = disco_actual.WriteAt([]byte(buf), newpos)
							if err != nil {
								msg_error(err)
							}
						}
						Consola += "Particion extendida creada con exito\n"

						// Se crea la particion implementando el mejor ajuste
					} else if string(masterboot.Mbr_disk_fit[:]) == "B" {
						bestIndex := numPartition
						for i := 0; i < 4; i++ {

							if byteToInt(masterboot.Mbr_partition[i].Part_start[:]) == -1 || (string(masterboot.Mbr_partition[i].Part_status[:]) == "1" && byteToInt(masterboot.Mbr_partition[i].Part_size[:]) >= size_bytes) {
								if i != numPartition {
									if byteToInt(masterboot.Mbr_partition[bestIndex].Part_size[:]) > byteToInt(masterboot.Mbr_partition[i].Part_size[:]) {
										bestIndex = i
										break
									}
								}
							}
						}

						copy(masterboot.Mbr_partition[bestIndex].Part_type[:], "E")
						copy(masterboot.Mbr_partition[bestIndex].Part_fit[:], auxFit)

						if bestIndex == 0 {
							var sizeMBR int = int(binary.Size(masterboot))
							copy(masterboot.Mbr_partition[bestIndex].Part_start[:], strconv.Itoa(sizeMBR))
						} else {
							var part_start int = byteToInt(masterboot.Mbr_partition[bestIndex-1].Part_start[:]) + byteToInt(masterboot.Mbr_partition[bestIndex-1].Part_size[:])
							copy(masterboot.Mbr_partition[bestIndex].Part_start[:], strconv.Itoa(part_start))
						}

						copy(masterboot.Mbr_partition[bestIndex].Part_size[:], strconv.Itoa(size_bytes))
						copy(masterboot.Mbr_partition[bestIndex].Part_status[:], "0")
						copy(masterboot.Mbr_partition[bestIndex].Part_name[:], nombreP)

						// Se guarda el MBR actualizado
						disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
						s1 := &masterboot
						// Escribimos struct.
						var binario3 bytes.Buffer
						binary.Write(&binario3, binary.BigEndian, s1)
						escribirDentroFILE(disco_actual, binario3.Bytes())

						// Se posiciona en el part_start de la nueva particion en el disco
						var ppart_start int = byteToInt(masterboot.Mbr_partition[bestIndex].Part_start[:])
						disco_actual.Seek(int64(ppart_start), 0)
						// Se guarda la particion extendida
						extendedBoot := strct.EBR{}
						copy(extendedBoot.Part_fit[:], auxFit)
						copy(extendedBoot.Part_status[:], "0")
						copy(extendedBoot.Part_start[:], strconv.Itoa(byteToInt(masterboot.Mbr_partition[bestIndex].Part_start[:])))
						copy(extendedBoot.Part_size[:], strconv.Itoa(0))
						copy(extendedBoot.Part_next[:], strconv.Itoa(-1))
						copy(extendedBoot.Part_name[:], "")

						// Se escribe el EBR en el disco, en el part_start de la particion ubicado ya con fseek
						s2 := &extendedBoot
						// Escribimos struct.
						var binario4 bytes.Buffer
						binary.Write(&binario4, binary.BigEndian, s2)
						escribirDentroFILE(disco_actual, binario4.Bytes())

						// Se escribe 1 despues del EBR, indicando donde comienza la particion
						//var sizeEBR int = int(binary.Size(extendedBoot))
						for k := 0; k < 1; k++ {
							// Cambio de posicion de puntero dentro del archivo
							newpos, err := disco_actual.Seek(int64(k+ppart_start), 0)
							if err != nil {
								msg_error(err)
							}
							// Escritura de struct en archivo binario
							_, err = disco_actual.WriteAt([]byte(buf), newpos)
							if err != nil {
								msg_error(err)
							}
						}
						Consola += "Particion extendida creada con exito\n"

						// Se crea la particion implementando el peor ajuste
					} else if string(masterboot.Mbr_disk_fit[:]) == "W" {

						worstIndex := numPartition
						for i := 0; i < 4; i++ {
							if byteToInt(masterboot.Mbr_partition[i].Part_start[:]) == -1 || (string(masterboot.Mbr_partition[i].Part_status[:]) == "1" && byteToInt(masterboot.Mbr_partition[i].Part_size[:]) >= size_bytes) {
								if i != numPartition {
									if byteToInt(masterboot.Mbr_partition[worstIndex].Part_size[:]) < byteToInt(masterboot.Mbr_partition[i].Part_size[:]) {
										worstIndex = i
										break
									}
								}
							}
						}

						copy(masterboot.Mbr_partition[worstIndex].Part_type[:], "E")
						copy(masterboot.Mbr_partition[worstIndex].Part_fit[:], auxFit)

						if worstIndex == 0 {
							var sizeMBR int = int(binary.Size(masterboot))
							copy(masterboot.Mbr_partition[worstIndex].Part_start[:], strconv.Itoa(sizeMBR))
						} else {
							var part_start int = byteToInt(masterboot.Mbr_partition[worstIndex-1].Part_start[:]) + byteToInt(masterboot.Mbr_partition[worstIndex-1].Part_size[:])
							copy(masterboot.Mbr_partition[worstIndex].Part_start[:], strconv.Itoa(part_start))
						}

						copy(masterboot.Mbr_partition[worstIndex].Part_size[:], strconv.Itoa(size_bytes))
						copy(masterboot.Mbr_partition[worstIndex].Part_status[:], "0")
						copy(masterboot.Mbr_partition[worstIndex].Part_name[:], nombreP)

						// Se guarda el MBR actualizado
						disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
						s1 := &masterboot
						// Escribimos struct.
						var binario3 bytes.Buffer
						binary.Write(&binario3, binary.BigEndian, s1)
						escribirDentroFILE(disco_actual, binario3.Bytes())

						// Se posiciona en el part_start de la nueva particion en el disco
						var ppart_start int = byteToInt(masterboot.Mbr_partition[worstIndex].Part_start[:])
						disco_actual.Seek(int64(ppart_start), 0)
						// Se guarda la particion extendida
						extendedBoot := strct.EBR{}
						copy(extendedBoot.Part_fit[:], auxFit)
						copy(extendedBoot.Part_status[:], "0")
						copy(extendedBoot.Part_start[:], strconv.Itoa(byteToInt(masterboot.Mbr_partition[worstIndex].Part_start[:])))
						copy(extendedBoot.Part_size[:], strconv.Itoa(0))
						copy(extendedBoot.Part_next[:], strconv.Itoa(-1))
						copy(extendedBoot.Part_name[:], "")

						// Se escribe el EBR en el disco, en el part_start de la particion ubicado ya con fseek
						s2 := &extendedBoot
						// Escribimos struct.
						var binario4 bytes.Buffer
						binary.Write(&binario4, binary.BigEndian, s2)
						escribirDentroFILE(disco_actual, binario4.Bytes())

						// Se escribe 1 despues del EBR, indicando donde comienza la particion
						//var sizeEBR int = int(binary.Size(extendedBoot))
						for k := 0; k < 1; k++ {
							// Cambio de posicion de puntero dentro del archivo
							newpos, err := disco_actual.Seek(int64(k+ppart_start), 0)
							if err != nil {
								msg_error(err)
							}
							// Escritura de struct en archivo binario
							_, err = disco_actual.WriteAt([]byte(buf), newpos)
							if err != nil {
								msg_error(err)
							}
						}
						Consola += "Particion extendida creada con exito\n"
					}
				} else {
					Consola += "ERROR: ya existe una particion con ese nombre\n"
				}

			} else {
				Consola += "ERROR: la particion a crear excede el tamano libre\n"
			}
		} else {
			Consola += "ERROR: Ya existen 4 particiones, no se puede crear otra\n"
			Consola += "Elimine alguna para poder crear una\n"
		}

	} else {
		Consola += "ERROR: ya existe una particion extendida en este disco\n"
	}
	disco_actual.Close()
}

func crearParticionLogica(tamano int, unidad string, path string, ajuste string, nombreP string) {
	auxFit := strings.ToUpper(ajuste[:1])
	size_bytes := 1024
	masterboot := strct.MBR{}

	if unidad == "m" {
		size_bytes = tamano * 1048576
	} else if unidad == "k" {
		size_bytes = tamano * 1024
	} else {
		// Tamaño en bytes
		size_bytes = tamano
	}

	// Si el disco es abierto con exito, se empieza a verificar si se puede crear la particion
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	numExtendida := -1 // Que numero de particion es

	// --------Se extrae el MBR del disco---------
	// Obtenemos el size del mbr
	var size int = int(binary.Size(masterboot))
	// Leemos la cantidad de bytes en el disco
	disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
	data := leerEnFILE(disco_actual, size)
	// Convierte la data en un buffer,necesario para
	// Decodificar binario
	buffer := bytes.NewBuffer(data)
	// Decodificamos y guardamos en la instancia del MBR
	err = binary.Read(buffer, binary.BigEndian, &masterboot)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// Con el MBR extraido, se busca si existe la particion extendida
	for i := 0; i < 4; i++ {
		if string(masterboot.Mbr_partition[i].Part_type[:]) == "E" {
			numExtendida = i
			break
		}
	}

	if !(existePartition(masterboot, nombreP)) {

		if numExtendida != -1 {
			extendedBoot := strct.EBR{}
			cont := byteToInt(masterboot.Mbr_partition[numExtendida].Part_start[:])

			// Se posiciona en el part_start del primer EBR
			// --------Se extrae el EBR del disco---------
			// Obtenemos el size del ebr
			var size int = int(binary.Size(extendedBoot))
			// Leemos la cantidad de bytes en el disco
			disco_actual.Seek(int64(cont), 0) // Nos posicionamos en el inicio del primer EBR
			data := leerEnFILE(disco_actual, size)
			// Convierte la data en un buffer,necesario para
			// Decodificar binario
			buffer := bytes.NewBuffer(data)
			// Decodificamos y guardamos en la instancia del EBR
			err = binary.Read(buffer, binary.BigEndian, &extendedBoot)
			if err != nil {
				Consola += "Binary.Read failed\n"
				msg_error(err)
			}

			// Si es la primera particion logica
			if byteToInt(extendedBoot.Part_size[:]) == 0 {

				// Verificacion si la nueva particion L cabe en la E
				if byteToInt(masterboot.Mbr_partition[numExtendida].Part_size[:]) < size_bytes {
					Consola += "ERROR. La particion logica a crear excede el espacio disponible de la particion extendida\n"
				} else {
					// Obtenemos la posicion actual del fichero
					newpos, err := disco_actual.Seek(0, os.SEEK_CUR)
					if err != nil {
						msg_error(err)
					}
					copy(extendedBoot.Part_fit[:], auxFit)
					copy(extendedBoot.Part_status[:], "0")
					copy(extendedBoot.Part_start[:], strconv.Itoa(int(newpos)-size)) //Para regresar al inicio de la extendida
					copy(extendedBoot.Part_size[:], strconv.Itoa(size_bytes))
					copy(extendedBoot.Part_next[:], strconv.Itoa(-1))
					copy(extendedBoot.Part_name[:], nombreP)

					// Se posiciona en el part_start de la nueva particion en el disco
					var ppart_start int = byteToInt(masterboot.Mbr_partition[numExtendida].Part_start[:])
					disco_actual.Seek(int64(ppart_start), 0)

					// Se escribe el EBR en el disco, en el part_start de la particion ubicado ya con fseek
					s2 := &extendedBoot
					// Escribimos struct.
					var binario4 bytes.Buffer
					binary.Write(&binario4, binary.BigEndian, s2)
					escribirDentroFILE(disco_actual, binario4.Bytes())

					Consola += "Particion logica creada con exito\n"
				}
			} else {
				// Si no es la primera particion Logica

				// Obtenemos la posicion actual del fichero
				newpos, err := disco_actual.Seek(0, os.SEEK_CUR)
				if err != nil {
					msg_error(err)
				}
				posActual := int(newpos)
				posExtendida := byteToInt(masterboot.Mbr_partition[numExtendida].Part_size[:]) + byteToInt(masterboot.Mbr_partition[numExtendida].Part_start[:])
				for (byteToInt(extendedBoot.Part_next[:]) != -1) && (posActual < posExtendida) {
					var ebrPartNext int = byteToInt(extendedBoot.Part_next[:])
					disco_actual.Seek(int64(ebrPartNext), 0)

					// Se sobreescribe el siguiente EBR
					data := leerEnFILE(disco_actual, size)
					// Convierte la data en un buffer,necesario para
					// Decodificar binario
					buffer := bytes.NewBuffer(data)
					// Decodificamos y guardamos en la instancia del EBR
					err = binary.Read(buffer, binary.BigEndian, &extendedBoot)
					if err != nil {
						Consola += "Binary.Read failed\n"
						msg_error(err)
					}

					// Obtenemos la posicion actual del fichero
					curpos, err := disco_actual.Seek(0, os.SEEK_CUR)
					if err != nil {
						msg_error(err)
					}
					posActual = int(curpos)
				}

				espacioNecesario := byteToInt(extendedBoot.Part_start[:]) + byteToInt(extendedBoot.Part_size[:]) + size_bytes
				if espacioNecesario <= posExtendida {

					part_nexNew := byteToInt(extendedBoot.Part_start[:]) + byteToInt(extendedBoot.Part_size[:])
					copy(extendedBoot.Part_next[:], strconv.Itoa(part_nexNew))
					// Escribimos el next del ultimo EBR
					// Obtenemos la posicion actual del fichero
					curpos, err := disco_actual.Seek(0, os.SEEK_CUR)
					if err != nil {
						msg_error(err)
					}
					disco_actual.Seek(curpos-int64(size), 0)

					// Se escribe el EBR en el disco, en el part_start de la particion ubicado ya con fseek
					s2 := &extendedBoot
					// Escribimos struct.
					var binario4 bytes.Buffer
					binary.Write(&binario4, binary.BigEndian, s2)
					escribirDentroFILE(disco_actual, binario4.Bytes())

					// Escribimos el nuevo EBR
					disco_actual.Seek(int64(part_nexNew), 0)
					// Obtenemos la posicion actual del fichero
					currPos, err := disco_actual.Seek(0, os.SEEK_CUR)
					if err != nil {
						msg_error(err)
					}
					newEBR := strct.EBR{}
					copy(newEBR.Part_fit[:], auxFit)
					copy(newEBR.Part_status[:], "0")
					copy(newEBR.Part_start[:], strconv.Itoa(int(currPos)))
					copy(newEBR.Part_size[:], strconv.Itoa(size_bytes))
					copy(newEBR.Part_next[:], strconv.Itoa(-1))
					copy(newEBR.Part_name[:], nombreP)

					// Se escribe el EBR en el disco, en el part_start de la particion ubicado ya con fseek
					s3 := &newEBR
					// Escribimos struct.
					var binario5 bytes.Buffer
					binary.Write(&binario5, binary.BigEndian, s3)
					escribirDentroFILE(disco_actual, binario5.Bytes())

					Consola += "Particion logica creada con exito\n"

				} else {
					Consola += "ERROR la particion logica a crear excede el\n"
					Consola += "espacio disponible de la particion extendida\n"
				}
			}
		} else {
			Consola += "ERROR se necesita una particion extendida donde guardar la logica\n"
		}
	} else {
		Consola += "ERROR ya existe una particion con ese nombre\n"
	}
	disco_actual.Close()
}

func byteToStr(array []byte) string { //paso de []byte a string (SIRVE EN ESPECIAL PARA OBTENER UN VALOR NUMERICO)
	contador := 0
	str := ""
	for {
		if contador == len(array) { //significa que termine de recorrel el array
			break
		} else {
			//if array[contador] == uint8(56) { //no hago nada
			//str += "0"
			//}
			if array[contador] == uint8(0) {
				array[contador] = uint8(0) //asigno \00 (creo) y finalizo
				break
			} else if array[contador] != 0 {
				str += string(array[contador]) //le agrego a mi cadena un valor real
			}
		}
		contador++
	}

	return str
}

func byteToInt(part []byte) int {

	fus := -7777777777777777777 //numero malo xd

	ff1 := byteToStr(part[:]) //el tam del DD q posee el mbr lo obtengo en string
	//fmt.Println("/////desencadenando     " + ff1)
	partSize, err := strconv.Atoi(ff1) //valor string lo convierto a int
	//comprobamos q no exista error
	if err != nil {
		msg_error(err)
		return fus
	}
	fus = partSize

	return fus
}

func leerEnFILE(file *os.File, n int) []byte { //leemos n bytes del DD y lo devolvemos
	Arraybytes := make([]byte, n)   //molde q contendra lo q leemos
	_, err := file.Read(Arraybytes) // recogemos la info q nos interesa y la guardamos en el molde

	if err != nil { //si es error lo reportamos
		msg_error(err)
	}
	return Arraybytes
}

func escribirDentroFILE(file *os.File, bytes []byte) { //escribe dentro de un file
	_, err := file.Write(bytes)

	if err != nil {
		msg_error(err)
	}
}

func existePartition(mbr strct.MBR, name string) bool {
	pos_extendida := -1

	for i := 0; i < 4; i++ {

		if strings.Compare(byteToStr(mbr.Mbr_partition[i].Part_name[:]), name) == 0 {
			return true
		} else if string(mbr.Mbr_partition[i].Part_type[:]) == "e" {
			pos_extendida = i
		}
	}

	if pos_extendida != -1 {
		// FALTO VERIFICAR SI EL NOMBRE EXISTE EN LAS LOGICAS
		return false
	}

	return false
}

func show_Particiones(path string) {
	masterboot := strct.MBR{}

	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	// --------Se extrae el MBR del disco---------
	var size int = int(binary.Size(masterboot))
	disco_actual.Seek(0, 0) // Nos posicionamos en el inicio del archivo.
	data := leerEnFILE(disco_actual, size)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &masterboot)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	Consola += "<<------------------------ PARTICIONES ---------------------------->>\n"
	for i := 0; i < 4; i++ {
		if string(masterboot.Mbr_partition[i].Part_status[:]) != "1" {
			Consola += "<< ------------------- " + strconv.Itoa(i) + " ------------------->>\n"
			Consola += "Estado: " + byteToStr(masterboot.Mbr_partition[i].Part_status[:]) + "\n"
			Consola += "Nombre: " + byteToStr(masterboot.Mbr_partition[i].Part_name[:]) + "\n"
			Consola += "Fit: " + byteToStr(masterboot.Mbr_partition[i].Part_fit[:]) + "\n"
			Consola += "Part_start: " + byteToStr(masterboot.Mbr_partition[i].Part_start[:]) + "\n"
			Consola += "Size: " + byteToStr(masterboot.Mbr_partition[i].Part_size[:]) + "\n"
			Consola += "Type: " + byteToStr(masterboot.Mbr_partition[i].Part_type[:]) + "\n"

			if string(masterboot.Mbr_partition[i].Part_type[:]) == "E" {
				logicaR := strct.EBR{}
				var epart_start int = byteToInt(masterboot.Mbr_partition[i].Part_start[:])
				disco_actual.Seek(int64(epart_start), 0)

				var esize int = int(binary.Size(logicaR))
				data := leerEnFILE(disco_actual, esize)
				buffer := bytes.NewBuffer(data)
				err = binary.Read(buffer, binary.BigEndian, &logicaR)
				if err != nil {
					Consola += "Binary.Read failed\n"
					msg_error(err)
				}

				for byteToInt(logicaR.Part_next[:]) != -1 {
					Consola += "<< -------------------- Particion Logica --------------------->>\n"
					Consola += "Nombre: " + byteToStr(logicaR.Part_name[:]) + "\n"
					Consola += "Fit: " + byteToStr(logicaR.Part_fit[:]) + "\n"
					Consola += "Part_start: " + byteToStr(logicaR.Part_start[:]) + "\n"
					Consola += "Size: " + byteToStr(logicaR.Part_size[:]) + "\n"
					Consola += "Part_next: " + byteToStr(logicaR.Part_next[:]) + "\n"
					Consola += "Estado: " + byteToStr(logicaR.Part_status[:]) + "\n"

					disco_actual.Seek(int64(byteToInt(logicaR.Part_next[:])), 0)
					data := leerEnFILE(disco_actual, esize)
					buffer := bytes.NewBuffer(data)
					err = binary.Read(buffer, binary.BigEndian, &logicaR)
					if err != nil {
						Consola += "Binary.Read failed\n"
						msg_error(err)
					}

				}
			}
		}
	}
	disco_actual.Close()
}
