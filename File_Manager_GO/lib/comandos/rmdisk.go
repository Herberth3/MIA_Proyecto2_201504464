package comandos

import (
	"os"
)

func EjecutarRMDISK(path string) {

	// Se limpia la cadena Consola para recolectar informacion de una nueva ejecucion
	Consola = ""
	err := os.Remove(path)
	// Por si tiene error
	if err != nil {
		msg_error(err)
	} else {
		Consola += "Disco Duro eliminado!\n"
	}
}
