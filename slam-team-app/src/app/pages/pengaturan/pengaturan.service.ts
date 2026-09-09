import { Injectable, inject } from '@angular/core';

import { ApiService } from '../../core/services/api.service';
import { SettingGroup, SettingKV } from './pengaturan.model';

@Injectable({ providedIn: 'root' })
export class PengaturanService {
  private api = inject(ApiService);

  list() {
    return this.api.get<SettingGroup[]>('/pengaturan');
  }

  bulkUpdate(items: SettingKV[]) {
    return this.api.put<SettingGroup[]>('/pengaturan', { items });
  }
}
