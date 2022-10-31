import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ServicioService {

  constructor(private httpClient: HttpClient) { }

  postEntrada(entrada: string){
    return this.httpClient.post("http://localhost:5000/analizar",{ Cmd: entrada});
  }
}
