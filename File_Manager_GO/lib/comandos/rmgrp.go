package comandos

import (
	strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"unsafe"
)

func EjecutarRMGRP(grp_name string) {
	Consola = ""

	if IsLoginFlag {

		// Validacion que sea el Usuario ROOT
		if CurrentSession.Id_user == 1 && CurrentSession.Id_grp == 1 {

			// Validacion si el grupo ya existe
			// Metodo buscarGrupo() pertenece al comando Login.go
			grp_id := buscarGrupo(grp_name)
			if grp_id != -1 {
				eliminarGrupo(grp_name)

			} else {
				Consola += "Error: El grupo no existe!\n"
			}

		} else {
			Consola += "Error: Solo el usuario root puede ejecutar este comando!\n"
		}

	} else {
		Consola += "Error: Necesita iniciar sesion para poder ejecutar este comando!\n"
	}
}

/* Metodo para eliminar un grupo del archivo users.txt de una particion
 * @param string grp_name: Nombre del grupo a eliminar
 */
func eliminarGrupo(grp_name string) {
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
	archivo := strct.BloqueArchivo{}
	archivo1 := strct.BloqueArchivo{}

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

	col := 1
	actual := ""
	posicion := 0
	numBloque := 0
	//id := -1
	tipo := "*"
	grupo := ""
	palabra := ""
	flag := false

	for i := 0; i < 12; i++ {

		// El 255 representa al -1
		if int(inodo.I_block[i]) != 255 {

			disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*int(inodo.I_block[i])), 0)
			// --------Se extrae un Bloque Archivo---------
			var basize int = int(binary.Size(archivo))
			dataA := leerEnFILE(disco_actual, basize)
			buff := bytes.NewBuffer(dataA)
			err = binary.Read(buff, binary.BigEndian, &archivo)
			if err != nil {
				Consola += "Binary.Read failed\n"
				msg_error(err)
			}

			for j := 0; j < 63; j++ {

				actual = string(archivo.B_content[j])
				if actual == "\n" {
					if tipo == "G" {
						grupo = palabra

						if strings.Compare(grupo, grp_name) == 0 {

							disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*numBloque), 0)
							// --------Se extrae un nuevo Bloque Archivo---------
							var basize int = int(binary.Size(archivo1))
							dataA := leerEnFILE(disco_actual, basize)
							buff := bytes.NewBuffer(dataA)
							err = binary.Read(buff, binary.BigEndian, &archivo1)
							if err != nil {
								Consola += "Binary.Read failed\n"
								msg_error(err)
							}

							zeroString := "0"
							archivo1.B_content[posicion] = zeroString[0]

							// Almacenar el bloque archivo modificado
							disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*numBloque), 0)
							sBloqueA := &archivo1
							var binario1 bytes.Buffer
							binary.Write(&binario1, binary.BigEndian, sBloqueA)
							escribirDentroFILE(disco_actual, binario1.Bytes())

							Consola += "Grupo eliminado con exito!\n"
							flag = true
							break

						}

					}
					col = 1
					palabra = ""

				} else if actual != "," {
					palabra += actual
					col++

				} else if actual == "," {
					if col == 2 {
						//id, _ = strconv.Atoi(palabra)
						posicion = j - 1
						numBloque = int(inodo.I_block[i])
					} else if col == 4 {
						tipo = string(palabra[0])
					}
					col++
					palabra = ""

				}
			}

			if flag {
				break
			}
		}
	}

	disco_actual.Close()
}
