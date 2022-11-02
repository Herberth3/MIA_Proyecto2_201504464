import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ServicioService {

  constructor(private httpClient: HttpClient) { }

  postEntrada(entrada: string){
    return this.httpClient.post("http://54.205.237.197:5000/analizar",{ Cmd: entrada});
  }

  getUser(){
    return this.httpClient.get("http://54.205.237.197:5000/usuario");
  }
}
