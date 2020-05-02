import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import {TelemetryComponent} from "./page/telemetry/telemetry.component";


const routes: Routes = [
  {path: '', component: TelemetryComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
