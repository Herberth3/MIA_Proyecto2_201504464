import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { InterpreteComponent } from './components/interprete/interprete.component';
import { LoginComponent } from './components/login/login.component';

/**
 * Estableciendo rutas para los componentes creados por nosotros
 */
const routes: Routes = [
  { path: 'interprete', component: InterpreteComponent},
  { path: 'login', component: LoginComponent},
  { path: '**', redirectTo: 'interprete' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
