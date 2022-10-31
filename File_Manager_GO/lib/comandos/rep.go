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

// Variables que almacenaran el codigo .dot para ejecutarlo en graphviz desde el frontend
var RepDot string

func EjecutarREP(name string, id string) {

	// Se limpia la cadena Consola para recolectar informacion de una nueva ejecucion
	Consola = ""
	RepDot = ""

	error_ := 0
	// Para reporte MBR, DISK, SB, TREE
	pathDisco_Partition := ""

	// EL REPORTE SB, TREE, SE IMPLEMENTO SOLO PARA PARTICIONES PRIMARIAS
	// Para reporte SB, TREE
	startPartition := 0
	// sizePartition no se utiliza, se creo solo por la referencia en getDatosID
	sizePartition := 0
	// Para reporte SB
	nombre_disco := ""

	getDatosID(id, &pathDisco_Partition, &startPartition, &sizePartition, &error_)

	// Si 'error_' obtiene un 1, el path del ID montado no existe
	if error_ == 1 {
		return
	}

	// SI EL DISCO EXISTE, SE VERIFICA QUE REPORTE EJECUTAR
	if name == "disk" {
		graficarDisco(pathDisco_Partition)
	} else if name == "sb" {
		index_last_slash := strings.LastIndex(pathDisco_Partition, "/")
		nombre_disco = pathDisco_Partition[index_last_slash+1:]

		graficarSB(pathDisco_Partition, nombre_disco, startPartition)
	} else if name == "tree" {
		graficarTREE(pathDisco_Partition, startPartition)
	} else if name == "file" {
		// Falta file
		Consola += "File no implementado"
	} else {
		Consola += "Error: Nombre de reporte incorrecto\n"
	}
}

/*
Metodo para graficar un disco con la estructura de las particiones

	@param string path: Es el directorio donde se encuentra la particion
*/
func graficarDisco(path string) {
	// Se limpia el strings que almacena el codigo dot del reporte
	RepDot = ""

	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	// Extraccion de MBR
	mbr_auxiliar := strct.MBR{}
	// --------Se extrae el MBR del disco---------
	var size int = int(binary.Size(mbr_auxiliar))
	disco_actual.Seek(0, 0)
	data := leerEnFILE(disco_actual, size)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &mbr_auxiliar)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	total := byteToInt(mbr_auxiliar.Mbr_size[:])

	RepDot += "digraph G{\n\n"
	RepDot += "  tbl [\n    shape=box\n    label=<\n"
	RepDot += "     <table border=\"0\" cellborder=\"2\" width=\"600\" height=\"200\" color=\"LIGHTSTEELBLUE\">\n"
	RepDot += "     <tr>\n"
	RepDot += "     <td height=\"200\" width=\"100\"> MBR </td>\n"

	for i := 0; i < 4; i++ {

		parcial := byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:])

		if byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:]) != -1 { // Particion vacia

			porcentaje_real := float64(parcial) / (float64(total) / 100)
			porcentaje_aux := (float64(porcentaje_real) * 5)

			if string(mbr_auxiliar.Mbr_partition[i].Part_status[:]) != "1" {
				if string(mbr_auxiliar.Mbr_partition[i].Part_type[:]) == "P" {

					RepDot += "     <td height=\"200\" width=\"" + strconv.FormatFloat(porcentaje_aux, 'f', 1, 64) + "\">PRIMARIA <br/> Ocupado: " + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "%</td>\n"

					// Verificar que no haya espacio fragmentado
					if i != 3 {
						p1 := byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:]) + byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:])
						p2 := byteToInt(mbr_auxiliar.Mbr_partition[i+1].Part_start[:])

						if byteToInt(mbr_auxiliar.Mbr_partition[i+1].Part_start[:]) != -1 {

							if (p2 - p1) != 0 { // Hay fragmentacion
								fragmentacion := p2 - p1
								porciento_real := (float64(fragmentacion) * 100) / float64(total)
								porciento_aux := (float64(porciento_real) * 500) / 100

								RepDot += "     <td height=\"200\" width=\"" + strconv.FormatFloat(porciento_aux, 'f', 1, 64) + "\">LIBRE<br/> Ocupado: " + strconv.FormatFloat(porciento_real, 'f', 1, 64) + "%</td>\n"
							}
						}
					} else {
						p1 := byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:]) + byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:])
						mbr_size := total + size // size es el tamaño del struct MBR

						if (mbr_size - p1) != 0 { // Libre
							libre := (float64(mbr_size) - float64(p1)) + float64(size)
							porcent_real := (float64(libre) * 100) / float64(total)
							porcent_aux := (float64(porcent_real) * 500) / 100

							RepDot += "     <td height=\"200\" width=\"" + strconv.FormatFloat(porcent_aux, 'f', 1, 64) + "\">LIBRE<br/> Ocupado: " + strconv.FormatFloat(porcent_real, 'f', 1, 64) + "%</td>\n"
						}
					}
				} else { // Extendida
					extendedBoot := strct.EBR{}
					RepDot += "     <td  height=\"200\" width=\"" + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "\">\n     <table border=\"0\"  height=\"200\" WIDTH=\"" + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "\" cellborder=\"1\">\n"
					RepDot += "     <tr>  <td height=\"60\" colspan=\"15\">EXTENDIDA</td>  </tr>\n     <tr>\n"

					// --------Se extrae el primer EBR del disco---------
					var ebrsize int = int(binary.Size(extendedBoot))
					disco_actual.Seek(int64(byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:])), 0)
					data := leerEnFILE(disco_actual, ebrsize)
					buffer := bytes.NewBuffer(data)
					err = binary.Read(buffer, binary.BigEndian, &extendedBoot)
					if err != nil {
						Consola += "Binary.Read failed\n"
						msg_error(err)
					}

					if byteToInt(extendedBoot.Part_size[:]) != 0 { // Por si hay mas de alguna logica
						disco_actual.Seek(int64(byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:])), 0)

						// Obtenemos la posicion actual del fichero
						newpos, err := disco_actual.Seek(0, os.SEEK_CUR)
						if err != nil {
							msg_error(err)
						}
						posActual := int(newpos)
						posExtendida := byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:]) + byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:])
						for posActual < posExtendida {
							// --------Se extrae el primer EBR del disco para verificar que exista---------
							data := leerEnFILE(disco_actual, ebrsize)
							buffer := bytes.NewBuffer(data)
							err = binary.Read(buffer, binary.BigEndian, &extendedBoot)
							if err != nil {
								Consola += "Binary.Read failed, EBR REP DISK, ALL OK\n"
								msg_error(err)
								break
							}

							parcial = byteToInt(extendedBoot.Part_size[:])
							porcentaje_real = (float64(parcial) * 100) / float64(total)

							if porcentaje_real != 0 {
								if string(extendedBoot.Part_status[:]) != "1" {
									RepDot += "     <td height=\"140\">EBR</td>\n"
									RepDot += "     <td height=\"140\">LOGICA<br/>Ocupado: " + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "%</td>\n"

								} else { // Espacio no asignado
									RepDot += "      <td height=\"150\">LIBRE 1 <br/> Ocupado: " + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "%</td>\n"
								}

								if byteToInt(extendedBoot.Part_next[:]) == -1 {
									parcial = (byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:]) + byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:])) - (byteToInt(extendedBoot.Part_start[:]) + byteToInt(extendedBoot.Part_size[:]))
									porcentaje_real = (float64(parcial) * 100) / float64(total)

									if porcentaje_real != 0 {
										RepDot += "     <td height=\"150\">LIBRE 2<br/> Ocupado: " + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "%</td>\n"
									}
									break
								} else {
									disco_actual.Seek(int64(byteToInt(extendedBoot.Part_next[:])), 0)

									// Obtenemos la posicion actual del fichero
									curpos, err := disco_actual.Seek(0, os.SEEK_CUR)
									if err != nil {
										msg_error(err)
									}
									posActual = int(curpos)
								}
							}
						}
					} else {
						RepDot += "     <td height=\"140\"> Ocupado " + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "%</td>"
					}

					RepDot += "     </tr>\n     </table>\n     </td>\n"
					// Verificar que no haya espacio fragmentado
					if i != 3 {
						p1 := byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:]) + byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:])
						p2 := byteToInt(mbr_auxiliar.Mbr_partition[i+1].Part_start[:])

						if byteToInt(mbr_auxiliar.Mbr_partition[i+1].Part_start[:]) != -1 {

							if (p2 - p1) != 0 { // Hay fragmentacion
								fragmentacion := p2 - p1
								porciento_real := (float64(fragmentacion) * 100) / float64(total)
								porciento_aux := (float64(porciento_real) * 500) / 100

								RepDot += "     <td height=\"200\" width=\"" + strconv.FormatFloat(porciento_aux, 'f', 1, 64) + "\">LIBRE<br/> Ocupado: " + strconv.FormatFloat(porciento_real, 'f', 1, 64) + "%</td>\n"
							}
						}
					} else {
						p1 := byteToInt(mbr_auxiliar.Mbr_partition[i].Part_start[:]) + byteToInt(mbr_auxiliar.Mbr_partition[i].Part_size[:])
						mbr_size := total + size // size es el tamaño del struct MBR

						if (mbr_size - p1) != 0 { // Libre
							libre := (float64(mbr_size) - float64(p1)) + float64(size)
							porcent_real := (float64(libre) * 100) / float64(total)
							porcent_aux := (float64(porcent_real) * 500) / 100

							RepDot += "     <td height=\"200\" width=\"" + strconv.FormatFloat(porcent_aux, 'f', 1, 64) + "\">LIBRE<br/> Ocupado: " + strconv.FormatFloat(porcent_real, 'f', 1, 64) + "%</td>\n"
						}
					}
				}
			} else { // Espacio no asignado
				RepDot += "     <td height=\"200\" width=\"" + strconv.FormatFloat(porcentaje_aux, 'f', 1, 64) + "\">LIBRE <br/> Ocupado: " + strconv.FormatFloat(porcentaje_real, 'f', 1, 64) + "%</td>\n"
			}
		}
	}

	RepDot += "     </tr> \n     </table>        \n>];\n\n}"

	disco_actual.Close()
	Consola += "Reporte Disco generado con exito!\n"
}

/*
Metodo para generar el reporte del SuperBloque de una particion

	@param string path: Es el directorio donde se encuentra la particion
	@param string name_disk: Nombre del disco donde se encuentra almacenado el SB
	@param int part_start_SB: Byte donde inicia el super bloque
*/
func graficarSB(path string, name_disk string, part_start_SB int) {
	// Se limpia el strings que almacena el codigo dot del reporte
	RepDot = ""

	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	// Extraccion de SB
	superB := strct.SuperBloque{}
	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(part_start_SB), 0)
	data := leerEnFILE(disco_actual, sbsize)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	RepDot += "digraph G{\n"
	RepDot += "    nodo [shape=none, fontname=\"Century Gothic\" label=<"
	RepDot += "   <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" bgcolor=\"cornflowerblue\">"
	RepDot += "    <tr> <td COLSPAN=\"2\"> <b>SUPERBLOQUE</b> </td></tr>\n"

	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> sb_nombre_hd </td> <td bgcolor=\"white\"> " + name_disk + " </td> </tr>\n"

	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_inodes_count </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_inodes_count[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_blocks_count </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_blocks_count[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_free_block_count </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_free_blocks_count[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_free_inodes_count </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_free_inodes_count[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_mtime </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_mtime[:]) + " </td></tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_umtime </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_umtime[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_mnt_count </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_mnt_count[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_magic </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_magic[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_inode_size </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_inode_size[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_block_size </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_block_size[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_first_ino </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_first_ino[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_first_blo </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_first_blo[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_bm_inode_start </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_bm_inode_start[:]) + " </td></tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_bm_block_start </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_bm_block_start[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_inode_start </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_inode_start[:]) + " </td> </tr>\n"
	RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> s_block_start </td> <td bgcolor=\"white\"> " + byteToStr(superB.S_block_start[:]) + " </td> </tr>\n"
	RepDot += "   </table>>]\n"
	RepDot += "\n}"

	disco_actual.Close()
	Consola += "Reporte SuperBloque generado con exito!\n"
}

func graficarTREE(path string, part_start_Partition int) {
	// Se limpia el strings que almacena el codigo dot del reporte
	RepDot = ""

	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		msg_error(err)
		return
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	superB := strct.SuperBloque{}
	inodo := strct.InodoTable{}
	carpeta := strct.BloqueCarpeta{}
	archivo := strct.BloqueArchivo{}
	apuntador := strct.BloqueApuntadores{}

	// Tamaño de algunas estructuras
	var inodoTable strct.InodoTable
	const i_size = unsafe.Sizeof(inodoTable)

	var blockCarpeta strct.BloqueCarpeta
	const bc_size = unsafe.Sizeof(blockCarpeta)

	var blockArchivo strct.BloqueArchivo
	const ba_size = unsafe.Sizeof(blockArchivo)

	var blockApuntador strct.BloqueApuntadores
	const bapu_size = unsafe.Sizeof(blockApuntador)

	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(part_start_Partition), 0)
	data := leerEnFILE(disco_actual, sbsize)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		Consola += "Binary.Read failed\n"
		msg_error(err)
	}

	aux := byteToInt(superB.S_bm_inode_start[:])
	i := 0

	RepDot += "digraph G{\n\n"
	RepDot += "    rankdir=\"LR\" \n"

	// Creamos lo inodos
	for aux < byteToInt(superB.S_bm_block_start[:]) {

		disco_actual.Seek(int64(byteToInt(superB.S_bm_inode_start[:])+i), 0)
		aux++
		port := 0
		dataBMI := getc(disco_actual)  // me devuelve el dato en byte
		bufINT := int(dataBMI)         // Lo convierto en int
		buffer := strconv.Itoa(bufINT) // Convierto el int a string

		if buffer == "1" {
			var inodosize int = int(binary.Size(inodo))
			disco_actual.Seek(int64(byteToInt(superB.S_inode_start[:])+int(i_size)*i), 0)
			data := leerEnFILE(disco_actual, inodosize)
			buffer := bytes.NewBuffer(data)
			err = binary.Read(buffer, binary.BigEndian, &inodo)
			if err != nil {
				Consola += "Binary.Read failed\n"
				msg_error(err)
			}

			RepDot += "    inodo_" + strconv.Itoa(i) + " [ shape=plaintext fontname=\"Century Gothic\" label=<\n"
			RepDot += "   <table bgcolor=\"royalblue\" border=\"0\" >"
			RepDot += "    <tr> <td colspan=\"2\"><b>Inode " + strconv.Itoa(i) + "</b></td></tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_uid </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_uid[:]) + " </td>  </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_gid </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_gid[:]) + " </td>  </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_size </td><td bgcolor=\"white\"> " + byteToStr(inodo.I_size[:]) + " </td> </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_atime </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_atime[:]) + " </td> </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_ctime </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_ctime[:]) + " </td> </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_mtime </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_mtime[:]) + " </td> </tr>\n"

			for b := 0; b < 15; b++ {
				RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_block_" + strconv.Itoa(port) + " </td> <td bgcolor=\"white\" port=\"f" + strconv.Itoa(b) + "\"> " + strconv.Itoa(int(inodo.I_block[b])) + " </td></tr>\n"
				port++
			}

			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_type </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_type[:]) + " </td>  </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_perm </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_perm[:]) + " </td>  </tr>\n"
			RepDot += "   </table>>]\n\n"

			// Creamos los bloques relacionados al inodo
			for j := 0; j < 15; j++ {
				port = 0

				// El 255 representa al -1
				if int(inodo.I_block[j]) != 255 {

					disco_actual.Seek(int64(byteToInt(superB.S_bm_block_start[:])+int(inodo.I_block[j])), 0)

					buffINT := int(getc(disco_actual))
					buffer := strconv.Itoa(buffINT)

					if buffer == "1" { // Bloque carpeta
						disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(bc_size)*int(inodo.I_block[j])), 0)
						// --------Se extrae el Bloque Carpeta del disco---------
						var bcsize int = int(binary.Size(carpeta))
						data := leerEnFILE(disco_actual, bcsize)
						buff := bytes.NewBuffer(data)
						err = binary.Read(buff, binary.BigEndian, &carpeta)
						if err != nil {
							Consola += "Binary.Read failed\n"
							msg_error(err)
						}

						RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
						RepDot += "   <table bgcolor=\"seagreen\" border=\"0\">\n"
						RepDot += "    <tr> <td colspan=\"2\"><b>Folder block " + strconv.Itoa(int(inodo.I_block[j])) + "</b></td></tr>\n"
						RepDot += "    <tr> <td bgcolor=\"mediumseagreen\"> b_name </td> <td bgcolor=\"mediumseagreen\"> b_inode </td></tr>\n"

						for c := 0; c < 4; c++ {
							RepDot += "    <tr> <td bgcolor=\"white\" > " + byteToStr(carpeta.B_content[c].B_name[:]) + " </td> <td bgcolor=\"white\"  port=\"f" + strconv.Itoa(port) + "\"> " + byteToStr(carpeta.B_content[c].B_inodo[:]) + " </td></tr>\n"
							port++
						}

						RepDot += "   </table>>]\n\n"

						// Relacion de bloques a inodos
						for c := 0; c < 4; c++ {
							if byteToInt(carpeta.B_content[c].B_inodo[:]) != -1 {

								if strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), ".") != 0 && strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), "..") != 0 {
									RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + ":f" + strconv.Itoa(c) + " -> inodo_" + byteToStr(carpeta.B_content[c].B_inodo[:]) + ";\n"
								}
							}
						}

					} else if buffer == "2" { // Bloque archivo
						disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*int(inodo.I_block[j])), 0)
						// --------Se extrae el Bloque Archivo del disco---------
						var basize int = int(binary.Size(archivo))
						data := leerEnFILE(disco_actual, basize)
						buff := bytes.NewBuffer(data)
						err = binary.Read(buff, binary.BigEndian, &archivo)
						if err != nil {
							Consola += "Binary.Read failed\n"
							msg_error(err)
						}

						RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
						RepDot += "   <table border=\"0\" bgcolor=\"sandybrown\">\n"
						RepDot += "    <tr> <td> <b>File block " + strconv.Itoa(int(inodo.I_block[j])) + "</b></td></tr>\n"
						RepDot += "    <tr> <td bgcolor=\"white\"> " + byteToStr(archivo.B_content[:]) + " </td></tr>\n"
						RepDot += "   </table>>]\n\n"

					} else if buffer == "3" { // Bloque apuntador
						disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(bapu_size)*int(inodo.I_block[j])), 0)
						// --------Se extrae el Bloque Apuntador del disco---------
						var bapusize int = int(binary.Size(apuntador))
						data := leerEnFILE(disco_actual, bapusize)
						buff := bytes.NewBuffer(data)
						err = binary.Read(buff, binary.BigEndian, &apuntador)
						if err != nil {
							Consola += "Binary.Read failed\n"
							msg_error(err)
						}

						RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
						RepDot += "   <table border=\"0\" bgcolor=\"khaki\">\n"
						RepDot += "    <tr> <td> <b>Pointer block " + strconv.Itoa(int(inodo.I_block[j])) + "</b></td></tr>\n"

						for a := 0; a < 16; a++ {
							RepDot += "    <tr> <td bgcolor=\"white\" port=\"f" + strconv.Itoa(port) + "\">" + string(apuntador.B_pointer[a]) + "</td> </tr>\n"
							port++
						}

						RepDot += "   </table>>]\n\n"

						// Bloques carpeta/archivo  del bloque de apuntadores
						for x := 0; x < 16; x++ {
							port = 0

							// El 255 representa al -1
							if int(apuntador.B_pointer[x]) != 255 {
								disco_actual.Seek(int64(byteToInt(superB.S_bm_block_start[:])+int(apuntador.B_pointer[x])), 0)

								buffINT := int(getc(disco_actual))
								buffer := strconv.Itoa(buffINT)

								if buffer == "1" {
									disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(bc_size)*int(apuntador.B_pointer[x])), 0)
									// --------Se extrae el Bloque Carpeta del disco---------
									var bcsize int = int(binary.Size(carpeta))
									data := leerEnFILE(disco_actual, bcsize)
									buff := bytes.NewBuffer(data)
									err = binary.Read(buff, binary.BigEndian, &carpeta)
									if err != nil {
										Consola += "Binary.Read failed\n"
										msg_error(err)
									}

									RepDot += "    bloque_" + string(apuntador.B_pointer[x]) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
									RepDot += "   <table bgcolor=\"seagreen\" border=\"0\">\n"
									RepDot += "    <tr> <td colspan=\"2\"><b>Folder block " + string(apuntador.B_pointer[x]) + "</b></td></tr>\n"
									RepDot += "    <tr> <td bgcolor=\"mediumseagreen\"> b_name </td> <td bgcolor=\"mediumseagreen\"> b_inode </td></tr>\n"

									for c := 0; c < 4; c++ {
										RepDot += "    <tr> <td bgcolor=\"white\" > " + byteToStr(carpeta.B_content[c].B_name[:]) + " </td> <td bgcolor=\"white\"  port=\"f" + strconv.Itoa(port) + "\"> " + byteToStr(carpeta.B_content[c].B_inodo[:]) + " </td></tr>\n"
										port++
									}

									RepDot += "   </table>>]\n\n"

									// Relacion de bloques a inodos
									for c := 0; c < 4; c++ {
										if byteToInt(carpeta.B_content[c].B_inodo[:]) != -1 {

											if strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), ".") != 0 && strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), "..") != 0 {
												RepDot += "    bloque_" + string(apuntador.B_pointer[x]) + ":f" + strconv.Itoa(c) + " -> inodo_" + byteToStr(carpeta.B_content[c].B_inodo[:]) + ";\n"
											}
										}
									}
								} else if buffer == "2" {
									disco_actual.Seek(int64(byteToInt(superB.S_block_start[:])+int(ba_size)*int(apuntador.B_pointer[x])), 0)
									// --------Se extrae el Bloque Archivo del disco---------
									var basize int = int(binary.Size(archivo))
									data := leerEnFILE(disco_actual, basize)
									buff := bytes.NewBuffer(data)
									err = binary.Read(buff, binary.BigEndian, &archivo)
									if err != nil {
										Consola += "Binary.Read failed\n"
										msg_error(err)
									}

									RepDot += "    bloque_" + string(apuntador.B_pointer[x]) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
									RepDot += "   <table border=\"0\" bgcolor=\"sandybrown\">\n"
									RepDot += "    <tr> <td> <b>File block " + string(apuntador.B_pointer[x]) + "</b></td></tr>\n"
									RepDot += "    <tr> <td bgcolor=\"white\"> " + byteToStr(archivo.B_content[:]) + " </td></tr>\n"
									RepDot += "   </table>>]\n\n"

								} else if buffer == "3" {
									// NO SE IMPLEMENTO
									Consola += ""
								}
							}
						}

						// Relacion de bloques apuntador a bloques archivos/carpetas
						for b := 0; b < 16; b++ {
							// El 255 representa al -1
							if int(apuntador.B_pointer[b]) != 255 {
								RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + ":f" + strconv.Itoa(b) + " -> bloque_" + string(apuntador.B_pointer[b]) + ";\n"
							}
						}
					}
					// Relacion de inodos a bloques
					RepDot += "    inodo_" + strconv.Itoa(i) + ":f" + strconv.Itoa(j) + " -> bloque_" + strconv.Itoa(int(inodo.I_block[j])) + "; \n"
				}
			}
		}
		i++

	}

	RepDot += "\n\n}"
	disco_actual.Close()

	Consola += "Reporte Tree generado con exito!\n"

}

/*
Metodo que lee un byte del archivo en la posicion en donde se encuentra el puntero
*/
func getc(f *os.File) byte {
	b := make([]byte, 1)
	_, err := f.Read(b)

	if err != nil { //si es error lo reportamos
		msg_error(err)
	}

	return b[0]
}
