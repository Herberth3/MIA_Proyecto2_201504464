package comandos

func EjecutarLOGOUT() {
	Consola = ""

	if IsLoginFlag {
		IsLoginFlag = false
		CurrentSession.Id_user = 0
		CurrentSession.Path = ""
		CurrentSession.User_name = ""
		CurrentSession.Start_SB = -1

		Consola += "Sesion finalizada!\n"
	} else {
		Consola += "Error: No hay ninguna sesion activa\n"
	}
}
