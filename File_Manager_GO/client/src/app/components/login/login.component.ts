import { Component, OnInit } from '@angular/core';
import { ServicioService } from 'src/app/services/servicio.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {

  idparticion = "";
  usuario = "";
  contrasena = "";
  isLogin: any;
  userName = "";
  consola = "";

  constructor(public service: ServicioService) { }

  ngOnInit(): void {
  }

  ejecutar() {
    const cmd = "login -usuario=" + this.usuario + " -password=" + this.contrasena + " -id=" + this.idparticion;
    const idP = this.idparticion;
    const usr = this.usuario;
    const pwd = this.contrasena;

    if (idP == "") {
      alert("Ingrese un ID para la particion!");
      return
    } else if (usr == "") {
      alert("Ingrese un usuario!");
      return
    } else if (pwd == "") {
      alert("Ingrese una contraseña!")
      return
    }

    if (cmd != "") {
      this.service.postEntrada(cmd).subscribe(async (res: any) => {
        this.consola = await res.Consola + "\n";
        this.isLogin = await res.IsLogin;
        this.userName = await res.LoginName;

        if (this.isLogin != 0 && this.userName != "") {
          alert("Inicio de sesion correcta!");
          this.idparticion = "";
          this.usuario = "";
          this.contrasena = "";
        } else {
          alert(this.consola);
        }
      });

    }
  }
}
