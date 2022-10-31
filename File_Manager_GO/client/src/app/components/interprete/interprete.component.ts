import { Component, OnInit } from '@angular/core';
import { graphviz } from 'd3-graphviz';
import { wasmFolder } from "@hpcc-js/wasm";
import { ServicioService } from 'src/app/services/servicio.service';

@Component({
  selector: 'app-interprete',
  templateUrl: './interprete.component.html',
  styleUrls: ['./interprete.component.css']
})
export class InterpreteComponent implements OnInit {

  entrada = "";
  salida = "";
  repDot = "";

  constructor(public service: ServicioService) { }

  ngOnInit(): void {
    
  }

  public async onFileSelected(event:any) {
    const file:File = event.target.files[0];
    this.entrada = await file.text();
  }

  ejecutar(){
    const cmd = this.entrada;
      if(cmd != ""){
        this.service.postEntrada(cmd).subscribe(async (res:any) => {
          this.salida = await res.Consola + "\n";
          this.repDot = await res.RepDot + "\n";

          if (this.repDot != "") {
            this.drawDot(this.repDot);
          }

        });
      }else alert("Ingrese texto para continuar!");

  }

  drawDot(dot: string) {
    // Por si vienen varios dots del mismo reporte. ##*## es el limitador que se le coloca en el backend
    var arrayDots = dot.split("##*##");
    for (let i = 0; i < arrayDots.length; i++) {
      if (arrayDots[i] != "") {
        // La union de '#rep i' sera el nombre del id del div en el html
        console.log(arrayDots[i]);
        wasmFolder('/assets/@hpcc-js/wasm/dist/');
        graphviz("#rep" + i).renderDot(arrayDots[i]);
      }
    }
  }

}
