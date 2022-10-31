package lib

import (
	cmd "File_Manager_GO/lib/comandos"
	"strconv"
	"strings"
)

// Struct que almacena la informacion necesaria en respuesta a la API
type colector struct {
	Salida  string
	IsLogin bool
	RepDot  string
}

// Instancia de colector tipo struct que recolectara informacion necesaria para respuesta de la API
var Recolector colector

func msg_error(err error) {
	Recolector.Salida += "Error: " + err.Error() + "\n"
}

func Analizar(inputText string) {

	// Atributo Salida de Recolector, almacenara el string que informa si hubo un error o el comando se ejecuto con exito
	Recolector.Salida = ""
	// En cada peticion crear un nuevo arreglo de 26 discos para montar
	cmd.Discos = [26]cmd.DISCOMONTADO{}
	// Se limpian los strings que almacenan el codigo dot de los reportes en cada ejecucion
	Recolector.RepDot = ""

	var arregloComandos []string = strings.Split(inputText, "\n")

	for i := 0; i < len(arregloComandos); i++ {

		// Validacion si la linea es un comentario
		if strings.HasPrefix(arregloComandos[i], "#") {
			Recolector.Salida += arregloComandos[i] + "\n"
			continue
		} else if strings.Compare(arregloComandos[i], "") == 0 {
			// Viene una linea en blanco
			continue
		}

		// Si la linea no es un comentario, se ejecuta el comando
		Recolector.Salida += arregloComandos[i] + "\n"
		ejecutar(arregloComandos[i])
		Recolector.Salida += "\n"
	}

}

func ejecutar(command string) {

	var parametros []string = strings.Split(command, " -")

	comando := strings.ToLower(parametros[0])

	size_valor := -1
	unit_valor := ""
	fit_valor := ""
	path_valor := ""
	type_valor := ""
	name_valor := ""
	id_valor := ""

	size_flag := 0
	unit_flag := 0
	fit_flag := 0
	path_flag := 0
	type_flag := 0
	name_flag := 0
	id_flag := 0

	for i := 1; i < len(parametros); i++ {

		parametro := strings.ToLower(parametros[i])

		switch comando {
		case "mkdisk":
			if strings.Contains(parametro, "size=") {

				if size_flag == 0 {
					valor := strings.Replace(parametro, "size=", "", 1)
					valorInt, err := strconv.Atoi(valor)

					if err != nil {
						msg_error(err)
						return
					}

					size_valor = valorInt
					size_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro SIZE repetido\n"
					return
				}
			} else if strings.Contains(parametro, "unit=") {

				if unit_flag == 0 {
					valor := strings.Replace(parametro, "unit=", "", 1)

					unit_valor = strings.ToLower(valor)
					unit_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro UNIT repetido\n"
					return
				}
			} else if strings.Contains(parametro, "fit=") {

				if fit_flag == 0 {
					valor := strings.Replace(parametro, "fit=", "", 1)

					fit_valor = strings.ToLower(valor)
					fit_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro FIT repetido\n"
					return
				}
			} else if strings.Contains(parametro, "path=") {

				if path_flag == 0 {
					// Se omite la variable "parametro" que contiene el parametro requerido -path
					// Se crea otra variable "paramPath" que contiene el parametro original, sin implementar toLower
					// Esto para que el valor (ruta) del "-path" sea el original, con mayusculas y minusculas
					paramPath := parametros[i]
					// Extraccion de subcadena, que tomara lo que viene despues de -path=
					valor := paramPath[5:]

					path_valor = valor
					path_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro PATH repetido\n"
					return
				}
			} else {
				Recolector.Salida += "Error: Parametro no permitido en MKDISK\n"
				return
			}
		case "rmdisk":
			if strings.Contains(parametro, "path=") {

				if path_flag == 0 {
					// Se omite la variable "parametro" que contiene el parametro requerido -path
					// Se crea otra variable "paramPath" que contiene el parametro original, sin implementar toLower
					// Esto para que el valor (ruta) del "-path" sea el original, con mayusculas y minusculas
					paramPath := parametros[i]
					// Extraccion de subcadena, que tomara lo que viene despues de -path=
					valor := paramPath[5:]

					path_valor = valor
					path_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro PATH repetido\n"
					return
				}
			} else {
				Recolector.Salida += "Error: Parametro no permitido en RMDISK\n"
				return
			}
		case "fdisk":
			if strings.Contains(parametro, "size=") {

				if size_flag == 0 {
					valor := strings.Replace(parametro, "size=", "", 1)
					valorInt, err := strconv.Atoi(valor)

					if err != nil {
						msg_error(err)
						return
					}

					size_valor = valorInt
					size_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro SIZE repetido\n"
					return
				}
			} else if strings.Contains(parametro, "unit=") {

				if unit_flag == 0 {
					valor := strings.Replace(parametro, "unit=", "", 1)

					unit_valor = strings.ToLower(valor)
					unit_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro UNIT repetido\n"
					return
				}
			} else if strings.Contains(parametro, "path=") {

				if path_flag == 0 {
					// Se omite la variable "parametro" que contiene el parametro requerido -path
					// Se crea otra variable "paramPath" que contiene el parametro original, sin implementar toLower
					// Esto para que el valor (ruta) del "-path" sea el original, con mayusculas y minusculas
					paramPath := parametros[i]
					// Extraccion de subcadena, que tomara lo que viene despues de -path=
					valor := paramPath[5:]

					path_valor = valor
					path_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro PATH repetido\n"
					return
				}
			} else if strings.Contains(parametro, "type=") {

				if type_flag == 0 {
					valor := strings.Replace(parametro, "type=", "", 1)

					type_valor = strings.ToLower(valor)
					type_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro TYPE repetido\n"
					return
				}
			} else if strings.Contains(parametro, "fit=") {

				if fit_flag == 0 {
					valor := strings.Replace(parametro, "fit=", "", 1)

					fit_valor = strings.ToLower(valor)
					fit_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro FIT repetido\n"
					return
				}
			} else if strings.Contains(parametro, "name=") {

				if name_flag == 0 {
					valor := strings.Replace(parametro, "name=", "", 1)

					name_valor = strings.ToLower(valor)
					name_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro NAME repetido\n"
					return
				}
			} else {
				Recolector.Salida += "Error: Parametro no permitido en FDISK\n"
				return
			}
		case "mount":
			if strings.Contains(parametro, "name=") {

				if name_flag == 0 {
					valor := strings.Replace(parametro, "name=", "", 1)

					name_valor = strings.ToLower(valor)
					name_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro NAME repetido\n"
					return
				}
			} else if strings.Contains(parametro, "path=") {

				if path_flag == 0 {
					// Se omite la variable "parametro" que contiene el parametro requerido -path
					// Se crea otra variable "paramPath" que contiene el parametro original, sin implementar toLower
					// Esto para que el valor (ruta) del "-path" sea el original, con mayusculas y minusculas
					paramPath := parametros[i]
					// Extraccion de subcadena, que tomara lo que viene despues de -path=
					valor := paramPath[5:]

					path_valor = valor
					path_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro PATH repetido\n"
					return
				}
			} else {
				Recolector.Salida += "Error: Parametro no permitido en MOUNT\n"
				return
			}
		case "mkfs":
			if strings.Contains(parametro, "id=") {

				if id_flag == 0 {
					valor := strings.Replace(parametro, "id=", "", 1)

					id_valor = strings.ToLower(valor)
					id_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro ID repetido\n"
					return
				}
			} else if strings.Contains(parametro, "type=") {

				if type_flag == 0 {
					valor := strings.Replace(parametro, "type=", "", 1)

					type_valor = strings.ToLower(valor)
					type_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro TYPE repetido\n"
					return
				}
			} else {
				Recolector.Salida += "Error: Parametro no permitido en MKFS\n"
				return
			}
		case "rep":
			if strings.Contains(parametro, "path=") {

				// NOTA: Para este proyecto solo se valida el path, pero no se utiliza para nada en este comando
				if path_flag == 0 {
					paramPath := parametros[i]
					valor := paramPath[5:]

					path_valor = valor
					path_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro PATH repetido\n"
					return
				}
			} else if strings.Contains(parametro, "name=") {

				if name_flag == 0 {
					valor := strings.Replace(parametro, "name=", "", 1)

					name_valor = strings.ToLower(valor)
					name_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro NAME repetido\n"
					return
				}
			} else if strings.Contains(parametro, "id=") {

				if id_flag == 0 {
					valor := strings.Replace(parametro, "id=", "", 1)

					id_valor = strings.ToLower(valor)
					id_flag = 1
				} else {
					Recolector.Salida += "Error: Parametro ID repetido\n"
					return
				}
			} else {
				Recolector.Salida += "Error: Parametro no permitido en REP\n"
				return
			}
		default:
			Recolector.Salida += "Comando no encontrado\n"
			return
		}
	}

	// Se extraen las comillas a los valores a continuacion
	path_valor = strings.Trim(path_valor, "\"")
	name_valor = strings.Trim(name_valor, "\"")

	switch comando {
	case "mkdisk":
		if size_valor == -1 {
			Recolector.Salida += "Error: Parametro SIZE no establecido\n"
			return
		} else if size_valor == 0 {
			Recolector.Salida += "Error: Parametro SIZE debe ser mayor a 0\n"
			return
		}

		if unit_flag == 1 {

			if unit_valor == "k" || unit_valor == "m" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en UNIT: " + unit_valor + "\n"
				return
			}
		} else {
			unit_valor = "m"
			unit_flag = 1
		}

		if fit_flag == 1 {

			if fit_valor == "ff" || fit_valor == "bf" || fit_valor == "wf" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en FIT: " + fit_valor + "\n"
				return
			}
		} else {
			fit_valor = "ff"
			fit_flag = 1
		}

		if path_flag == 0 {
			Recolector.Salida += "Error: Parametro PATH no establecido\n"
			return
		}

		if !(strings.HasSuffix(path_valor, ".dk")) {
			Recolector.Salida += "Error: Extension del disco no permitido\n"
			return
		}

		cmd.EjecutarMKDISK(size_valor, unit_valor, fit_valor, path_valor)
		Recolector.Salida += cmd.Consola

	case "rmdisk":
		if path_flag == 0 {
			Recolector.Salida += "Error: Parametro PATH no establecido\n"
			return
		}

		if !(strings.HasSuffix(path_valor, ".dk")) {
			Recolector.Salida += "Error: Extension del disco no permitido\n"
			return
		}

		cmd.EjecutarRMDISK(path_valor)
		Recolector.Salida += cmd.Consola

	case "fdisk":
		if size_valor == -1 {
			Recolector.Salida += "Error: Parametro SIZE no establecido\n"
			return
		} else if size_valor == 0 {
			Recolector.Salida += "Error: Parametro SIZE debe ser mayor a 0\n"
			return
		}

		if name_flag == 0 {
			Recolector.Salida += "Error: Parametro NAME no establecido\n"
			return
		}

		if unit_flag == 1 {

			if unit_valor == "k" || unit_valor == "m" || unit_valor == "b" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en UNIT: " + unit_valor + "\n"
				return
			}
		} else {
			unit_valor = "k"
			unit_flag = 1
		}

		if fit_flag == 1 {

			if fit_valor == "ff" || fit_valor == "bf" || fit_valor == "wf" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en FIT: " + fit_valor + "\n"
				return
			}
		} else {
			fit_valor = "wf"
			fit_flag = 1
		}

		if type_flag == 1 {
			if type_valor == "p" || type_valor == "e" || type_valor == "l" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en TYPE: " + type_valor + "\n"
				return
			}
		} else {
			type_valor = "p"
			type_flag = 1
		}

		if path_flag == 0 {
			Recolector.Salida += "Error: Parametro PATH no establecido\n"
			return
		}

		if !(strings.HasSuffix(path_valor, ".dk")) {
			Recolector.Salida += "Error: Extension del disco no permitido\n"
			return
		}

		cmd.EjecutarFDISK(size_valor, unit_valor, path_valor, type_valor, fit_valor, name_valor)
		Recolector.Salida += cmd.Consola

	case "mount":
		if name_flag == 0 {
			Recolector.Salida += "Error: Parametro NAME no establecido\n"
			return
		}

		if path_flag == 0 {
			Recolector.Salida += "Error: Parametro PATH no establecido\n"
			return
		}

		if !(strings.HasSuffix(path_valor, ".dk")) {
			Recolector.Salida += "Error: Extension del disco no permitido\n"
			return
		}

		cmd.EjecutarMOUNT(path_valor, name_valor)
		Recolector.Salida += cmd.Consola

	case "mkfs":
		if id_flag == 0 {
			Recolector.Salida += "Error: Parametro ID no establecido\n"
			return
		}

		if type_flag == 1 {
			if type_valor == "full" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en TYPE: " + type_valor + "\n"
				return
			}
		} else {
			type_valor = "full"
			type_flag = 1
		}

		if !(strings.HasPrefix(id_valor, "64")) {
			Recolector.Salida += "Error: El ID no cumple con la estructura requerida\n"
			return
		}

		cmd.EjecutarMKFS(id_valor, type_valor, "2fs")
		Recolector.Salida += cmd.Consola

	case "pause":
		// Para este proyecto no tiene accion
		return
	case "rep":
		if path_flag == 0 {
			Recolector.Salida += "Error: Parametro PATH no establecido\n"
			return
		}

		if name_flag == 0 {
			Recolector.Salida += "Error: Parametro NAME no establecido\n"
			return
		} else {

			if name_valor == "disk" || name_valor == "tree" || name_valor == "sb" || name_valor == "file" {

			} else {
				Recolector.Salida += "Error: Valor no permitido en NAME de REP: " + name_valor + "\n"
				return
			}
		}

		if id_flag == 0 {
			Recolector.Salida += "Error: Parametro ID no establecido\n"
			return
		}

		if !(strings.HasPrefix(id_valor, "64")) {
			Recolector.Salida += "Error: El ID no cumple con la estructura requerida\n"
			return
		}

		cmd.EjecutarREP(name_valor, id_valor)
		Recolector.Salida += cmd.Consola
		// Concatenacion del dot en cada reporte, se puede concatenar 2 o mas dot's del mismo reporte
		// Se usara un split en el front para separar los dot del mismo reporte
		// El simbolo "##*##" se utilizara como limitador para separar los dot del mismo reporte cuando se recorra en el front
		Recolector.RepDot += cmd.RepDot + "##*##"

	}
}
