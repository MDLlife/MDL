import {Component} from '@angular/core';
import {MatIconRegistry} from "@angular/material";
import 'rxjs/add/operator/takeWhile';

@Component({
    selector: 'app-root',
    templateUrl: './app.component.html',
    styleUrls: ['./app.component.scss']
})
export class AppComponent {

    constructor(iconRegistry: MatIconRegistry) {
        iconRegistry
            .registerFontClassAlias('fontawesome', 'fa')
            .registerFontClassAlias('mdl', 'mdl-icon');
    }

}
