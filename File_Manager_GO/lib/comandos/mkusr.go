package comandos

import (
	strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

func EjecutarMKUSR(usuario string, pwd string, grp string) {
	Consola = ""

	if IsLoginFlag {

		// Validacion que sea el Usuario ROOT
		if CurrentSession.Id_user == 1 && CurrentSession.Id_grp == 1 {

			// Validacion si el grupo ya existe
			// Metodo buscarGrupo() pertenece al comando Login.go
			grp_id := buscarGrupo(grp)
			if grp_id != -1 {

				exist_user := buscarUsuario(usuario)
				if !exist_user {
					// Falta corroborar el metodo  que crea el nuevo id
					newUsr_id := getNewUsr_id()
					nuevoUsuario := strconv.Itoa(newUsr_id) + ",U," + grp + "," + usuario + "," + pwd + "\n"
					setToFileUsersTxt(nuevoUsuario)
					Consola += "Usuario creado con exito!\n"
				} else {
					Consola += "Error: El usuario ya existe!\n"
				}

			} else {
				Consola += "Error: No se encuentra el grupo al que pertenecec el usuario!\n"
			}

		} else {
			Consola += "Error: Solo el usuario root puede ejecutar este comando!\n"
		}

	} else {
		Consola += "Error: Necesita iniciar sesion para poder ejecutar este comando!\n"
	}
}

/* Funcion para verificar la existencia de un usuario
 * @param string usr_name = nombre del usuario
 * @return true = existe | false = no existe
 */
func buscarUsuario(usr_name string) bool {

	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(CurrentSession.Path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return true
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

				// Verificar que sea tipo Usuario
				if strings.Compare(data[1], "U") == 0 {
					user := data[3]
					if strings.Compare(user, usr_name) == 0 {
						return true
					}

				}
			}
		}
	}

	return false
}

/* Funcion para obtener el id del nuevo usuario
 * @return id del ultimo usuario + 1
 */
func getNewUsr_id() int {
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

				// Verificar que sea tipo Usuario
				if strings.Compare(data[1], "U") == 0 {

					idU, _ := strconv.Atoi(data[0])
					id_auxiliar = idU

				}
			}
		}
	}

	// Se retorna el ultimo id del tipo Usuario encontrado sumandole 1
	return id_auxiliar + 1
}
