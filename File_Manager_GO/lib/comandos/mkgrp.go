package comandos

import (
	strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func EjecutarMKGRP(grp_name string) {
	Consola = ""

	if IsLoginFlag {

		// Validacion que sea el Usuario ROOT
		if CurrentSession.Id_user == 1 && CurrentSession.Id_grp == 1 {

			// Validacion si el grupo ya existe
			// Metodo buscarGrupo() pertenece al comando Login.go
			grp_id := buscarGrupo(grp_name)
			if grp_id == -1 {
				newGrp_id := getNewGrp_id()
				nuevoGrupo := strconv.Itoa(newGrp_id) + ",G," + grp_name + "\n"
				setToFileUsersTxt(nuevoGrupo)
				Consola += "Grupo creado con exito!\n"

			} else {
				Consola += "Error: Ya existe un grupo con ese nombre!\n"
			}

		} else {
			Consola += "Error: Solo el usuario root puede ejecutar este comando!\n"
		}

	} else {
		Consola += "Error: Necesita iniciar sesion para poder ejecutar este comando!\n"
	}
}

/*
Funcion para obtener un nuevo ID para el nuevo grupo
@return id del ultimo grupo +1
*/
func getNewGrp_id() int {
	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(CurrentSession.Path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return 0
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	superB := strct.SuperBloque{}
	inodo := strct.InodoTable{}

	// Tamaño de algunas estructuras
	var inodoTable strct.InodoTable
	const i_size = unsafe.Sizeof(inodoTable)

	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(CurrentSession.Start_SB), 0)
	data := leerEnFILE(disco_actual, sbsize)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// --------Se extrae el InodoTable del archivo user.txt---------
	var inodosize int = int(binary.Size(inodo))
	disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)), 0)
	dataI := leerEnFILE(disco_actual, inodosize)
	bufferI := bytes.NewBuffer(dataI)
	err = binary.Read(bufferI, binary.BigEndian, &inodo)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// Almacenara lo extraido del archivo
	contenidoFile := ""
	id_auxiliar := -1

	// Se recorren los iblock para conocer los punteros
	for i := 0; i < 15; i++ {

		// El 255 representa al -1
		if int(inodo.I_block[i]) != 255 {
			archivo := strct.BloqueArchivo{}
			disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])), 0)

			for j := 0; j <= int(inodo.I_block[i]); j++ {
				// --------Se extrae el Bloque Archivo USERS.TXT---------
				var basize int = int(binary.Size(archivo))
				data := leerEnFILE(disco_actual, basize)
				buff := bytes.NewBuffer(data)
				err = binary.Read(buff, binary.BigEndian, &archivo)
				if err != nil {
					Consola += "Binary.Read failed\n"
					msg_error(err)
				}
			}

			//Concatenar el contenido de cada bloque perteneciente al archivo users.txt
			contenidoFile += byteToStr(archivo.B_content[:])
		}
	}

	disco_actual.Close()

	var arregloU_G []string = strings.Split(contenidoFile, "\n")

	for _, filaU_G := range arregloU_G {

		// Se verifica que la fila obtenida del contenido no venga vacia
		if filaU_G != "" {
			var data []string = strings.Split(filaU_G, ",")

			// Verificar ID que no se un U/G eliminado
			if strings.Compare(data[0], "0") != 0 {

				// Verificar que sea tipo Grupo
				if strings.Compare(data[1], "G") == 0 {

					idG, _ := strconv.Atoi(data[0])
					id_auxiliar = idG

				}
			}
		}
	}

	// Se retorna el ultimo id del tipo Grupo encontrado sumandole 1
	return id_auxiliar + 1
}

/* Metodo para agregar un grupo/usuario al archivo users.txt de una particion
 * @param string newData: Datos del nuevo grupo/usuario
 */
func setToFileUsersTxt(newData string) {
	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(CurrentSession.Path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	superB := strct.SuperBloque{}
	inodo := strct.InodoTable{}
	inodo1 := strct.InodoTable{}
	archivo := strct.BloqueArchivo{}

	// Tamaño de algunas estructuras
	var inodoTable strct.InodoTable
	const i_size = unsafe.Sizeof(inodoTable)

	var ba strct.BloqueArchivo
	const ba_size = unsafe.Sizeof(ba)

	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(CurrentSession.Start_SB), 0)
	data := leerEnFILE(disco_actual, sbsize)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// --------Se extrae el InodoTable del archivo user.txt---------
	var inodosize int = int(binary.Size(inodo))
	disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)), 0)
	dataI := leerEnFILE(disco_actual, inodosize)
	bufferI := bytes.NewBuffer(dataI)
	err = binary.Read(bufferI, binary.BigEndian, &inodo)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	blockIndex := 0

	for i := 0; i < 12; i++ {

		// El 255 representa al -1
		if int(inodo.I_block[i]) != 255 {
			blockIndex = int(inodo.I_block[i]) // Indice del ultimo bloque utilizado del archivo
		}
	}

	disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*blockIndex), 0)
	// --------Se extrae un Bloque Archivo---------
	var basize int = int(binary.Size(archivo))
	dataA := leerEnFILE(disco_actual, basize)
	buff := bytes.NewBuffer(dataA)
	err = binary.Read(buff, binary.BigEndian, &archivo)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// Calcular si el bloque tipo archivo tiene aun espacio para escribir
	content := byteToStr(archivo.B_content[:])
	blockInUse := len(content)
	blockFree := 63 - blockInUse
	contador := 0
	escribir := len(newData)

	// Si la nueva data aun cabe en el bloque
	if escribir <= blockFree {
		// Escribir byte a byte la newData
		for i := 0; i < 64; i++ {

			if archivo.B_content[i] == 0 && contador < escribir {
				archivo.B_content[i] = newData[contador]

				if newData[contador] == '\n' {
					contador = escribir
					break
				}
				contador++
			}
		}

		// Almacenar el bloque archivo modificado
		disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*blockIndex), 0)
		sBloqueA := &archivo
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, sBloqueA)
		escribirDentroFILE(disco_actual, binario1.Bytes())

		disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)), 0)
		var inodosize1 int = int(binary.Size(inodo1))
		dataI1 := leerEnFILE(disco_actual, inodosize1)
		bufferI1 := bytes.NewBuffer(dataI1)
		err = binary.Read(bufferI1, binary.BigEndian, &inodo1)
		if err != nil {
			Consola += "Binary.Read failed\n"
			msg_error(err)
		}

		// Actualizacion del tamaño del inodo y la fecha
		copy(inodo1.I_size[:], strconv.Itoa(byteToInt(inodo1.I_size[:])+escribir))
		copy(inodo1.I_mtime[:], time.Now().String())

		// Almacenar el inodo actualizado
		disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)), 0)
		sInodo1 := &inodo1
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, sInodo1)
		escribirDentroFILE(disco_actual, binario2.Bytes())

	} else {
		aux := ""
		aux2 := ""
		// Este indice se actualiza con el recorrido de los 2 for siguientes
		i := 0

		// Este for no tiene indice nuevo, utiliza el 'i' declarado arriba
		for i <= blockFree {
			aux += string(newData[i])
			i++
		}

		for i < escribir {
			aux2 += string(newData[i])
			i++
		}

		// Guardamos lo que quepa en el primer bloque
		// Escribir byte a byte la newData
		for i := 0; i < 64; i++ {

			if archivo.B_content[i] == 0 && contador < len(aux) {
				archivo.B_content[i] = aux[contador]

				if aux[contador] == '\n' {
					contador = len(aux)
					break
				}
				contador++
			}
		}

		// Esto no hace nada, solo es para desaparecer
		if contador != 1 {
			Consola += ""
		}
		// La alerta de que 'contador' no se utiliza =D

		// Almacenar el bloque archivo modificado
		disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*blockIndex), 0)
		sBloqueA := &archivo
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, sBloqueA)
		escribirDentroFILE(disco_actual, binario1.Bytes())

		// Nuevo bloque archivo, se almacena el resto de la newData
		auxArchivo := strct.BloqueArchivo{}
		copy(auxArchivo.B_content[:], aux2)
		bit := buscarBit(disco_actual, "B", string(CurrentSession.Fit[:]))

		// Guardamos el bloque en el bitmap y en la tabla de bloques
		disco_actual.Seek(int64(byteToInt(superB.S_bm_block_start[:])+bit), 0)
		Fputc('2', disco_actual)

		disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*bit), 0)
		sBloqueA1 := &auxArchivo
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, sBloqueA1)
		escribirDentroFILE(disco_actual, binario2.Bytes())

		// Guardamos el modificado del inodo
		disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)), 0)
		var inodosize1 int = int(binary.Size(inodo1))
		dataI1 := leerEnFILE(disco_actual, inodosize1)
		bufferI1 := bytes.NewBuffer(dataI1)
		err = binary.Read(bufferI1, binary.BigEndian, &inodo1)
		if err != nil {
			Consola += "Binary.Read failed\n"
			msg_error(err)
		}

		// Actualizacion del tamaño del inodo y la fecha
		copy(inodo1.I_size[:], strconv.Itoa(byteToInt(inodo1.I_size[:])+escribir))
		copy(inodo1.I_mtime[:], time.Now().String())
		inodo1.I_block[blockIndex] = byte(bit)
		disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)), 0)
		sInodo1 := &inodo1
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, sInodo1)
		escribirDentroFILE(disco_actual, binario3.Bytes())

		// Guardamos la nueva cantidad de bloques libres y el primer bloque libre
		copy(superB.S_first_blo[:], strconv.Itoa(byteToInt(superB.S_first_blo[:])+1))
		copy(superB.S_free_blocks_count[:], strconv.Itoa(byteToInt(superB.S_free_blocks_count[:])-1))
		disco_actual.Seek(int64(CurrentSession.Start_SB), 0)
		//Se escribe el superbloque al inicio de la particion
		s1 := &superB
		var binario4 bytes.Buffer
		binary.Write(&binario4, binary.BigEndian, s1)
		escribirDentroFILE(disco_actual, binario4.Bytes()) //meto el superbloque en el inicio de la particion

	}
	disco_actual.Close()

}

/* Funcion que devuelve el bit libre en el bitmap de inodos/bloques segun el ajuste
 * @param FILE fp: archivo en el que se esta leyendo
 * @param string tipo: tipo de bit a buscar (Inodo/Bloque)
 * @param string fit: ajuste de la particion
 * @return -1 = Ya no existen bloques libres | # bit libre en el bitmap
 */
func buscarBit(file *os.File, tipo string, fit string) int {
	super := strct.SuperBloque{}
	inicio_bm := 0
	bit_libre := -1
	tam_bm := 0

	file.Seek(int64(CurrentSession.Start_SB), 0)
	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(super))
	data := leerEnFILE(file, sbsize)
	buffer := bytes.NewBuffer(data)
	err := binary.Read(buffer, binary.BigEndian, &super)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	// Si es inodo
	if tipo == "I" {
		tam_bm = byteToInt(super.S_inodes_count[:])
		inicio_bm = byteToInt(super.S_bm_inode_start[:])

		// Si es bloque
	} else if tipo == "B" {
		tam_bm = byteToInt(super.S_blocks_count[:])
		inicio_bm = byteToInt(super.S_bm_block_start[:])
	}

	//-------------- Tipo de ajuste a utilizar-------------------
	if fit == "F" { // Primer ajuste

		for i := 0; i < tam_bm; i++ {
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				bit_libre = i
				return bit_libre
			}
		}

		if bit_libre == -1 {
			return -1
		}

	} else if fit == "B" { // Mejor ajuste

		libres := 0
		auxLibres := -1

		for i := 0; i < tam_bm; i++ { // Primer recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					if auxLibres == -1 || auxLibres == 0 {
						auxLibres = libres
					} else {
						if auxLibres > libres {
							auxLibres = libres
						}
					}
					libres = 0
				}
			} else if buffer == "1" {
				if auxLibres == -1 || auxLibres == 0 {
					auxLibres = libres
				} else {
					if auxLibres > libres {
						auxLibres = libres
					}
				}
				libres = 0
			}
		}

		for i := 0; i < tam_bm; i++ { // Segundo recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					res := (i + 1) - libres
					return res
				}
			} else if buffer == "1" {
				if auxLibres == libres {
					res := (i + 1) - libres
					return res
				}
				libres = 0
			}
		}

		return -1
	} else if fit == "W" { // Peor ajuste
		libres := 0
		auxLibres := -1

		for i := 0; i < tam_bm; i++ { // Primer recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					if auxLibres == -1 || auxLibres == 0 {
						auxLibres = libres
					} else {
						if auxLibres < libres {
							auxLibres = libres
						}
					}
					libres = 0
				}
			} else if buffer == "1" {
				if auxLibres == -1 || auxLibres == 0 {
					auxLibres = libres
				} else {
					if auxLibres < libres {
						auxLibres = libres
					}
				}
				libres = 0
			}
		}

		for i := 0; i < tam_bm; i++ { // Segundo recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					res := (i + 1) - libres
					return res
				}
			} else if buffer == "1" {
				if auxLibres == libres {
					res := (i + 1) - libres
					return res
				}
				libres = 0
			}
		}

		return -1
	}

	return 0
}

// Fputc handles fputc().
//
// Writes a character to the stream and advances the position indicator.
//
// The character is written at the position indicated by the internal position
// indicator of the stream, which is then automatically advanced by one.
func Fputc(c int32, f *os.File) int32 {
	n, err := f.Write([]byte{byte(c)})
	if err != nil {
		return 0
	}

	return int32(n)
}
