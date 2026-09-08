# PROMPT — Build APP FileService + upload/preview (Fase 0)

Paste this whole prompt into Claude Code with the working directory at
`D:/xampp/htdocs/slam-team-apl/slam-team-app`.

---

## 0. FIRST — design method (mandatory)
Announce **"Using design-taste-frontend"**, run its audit, enforce the SLAM tokens in `D:/xampp/htdocs/slam-team-apl/workflow/_shared/design-tokens.md` (dark, SLAM red, UPPERCASE headings, red focus ring, 44px targets), map to Bootstrap 5. *If `blog-fe` restored, mirror it; else follow tokens.*

## 1. Read first
- `D:/xampp/htdocs/slam-team-apl/workflow/modules/02-file-management.md`.
- `D:/xampp/htdocs/slam-team-apl/workflow/_shared/conventions-app.md` — **§8 private images via the file endpoint** (the FileService pattern) + §3 ApiService.

## 2. Scope
A reusable **FileService** + upload/preview components every module reuses. Backend contract (envelope `{success,message,data,errors}`):

| Method | Path | Note |
|--------|------|------|
| POST | `/files` | multipart, returns `{uuid, variants, url}` |
| GET | `/files/:uuid/:varian` | authenticated blob (`original\|medium\|low`) |
| DELETE | `/files/:uuid` | soft delete |

## 3. Files to create under `src/app/shared/services/` + `shared/components/`
### `shared/services/file.service.ts` (`providedIn:'root'`) — conventions-app §8
```ts
@Injectable({ providedIn: 'root' })
export class FileService {
  private http = inject(HttpClient);
  private base = environment.apiURL;
  imageUrl(uuid: string, varian: 'original'|'medium'|'low' = 'medium') {
    return this.http.get(`${this.base}/files/${uuid}/${varian}`, { responseType: 'blob' })
      .pipe(map(b => URL.createObjectURL(b)));   // caller must revokeObjectURL
  }
  upload(field: string, file: File, extra?: Record<string,string>) {
    const fd = new FormData();
    fd.append(field, file);                        // DO NOT set Content-Type
    for (const [k,v] of Object.entries(extra ?? {})) fd.append(k, v);
    return this.http.post<ApiResponse<{uuid:string; id?:number}>>(`${this.base}/files`, fd).pipe(map(unwrap));
  }
}
```
### `shared/components/file-upload.ts/.html` — a dumb component
- `<input type="file">` → `FileService.upload('file', file, {kategori, reff_type})` → emits `{id, uuid}`; shows a preview from `imageUrl(uuid,'low')`; **revokes** the object URL on destroy / replace. Loading + error states. Emits the stored id for the parent form control.
### `shared/components/secure-image.ts` — an `<img>` that fetches a private variant as a blob
- Input `uuid` + `varian`; renders the blob object URL; placeholder when null; revokes on destroy.

## 4. Reuse rules
- Prefer `medium` in lists/cards, `low` for thumbnails, `original` only when needed.
- **Never** put a token-less `/files/...` URL in `<img src>` — private files need the blob fetch (interceptor adds the token).
- Do not re-implement HTTP/token logic — use `HttpClient` + the existing `authInterceptor`.

## 5. i18n
Add `COMMON.UPLOAD`, `COMMON.CHOOSE_FILE`, `COMMON.REMOVE`, `FILE.TOO_LARGE`, `FILE.PREVIEW` to BOTH IND & ENG.

## 6. Verification checklist
- [ ] `npm run build:local` compiles.
- [ ] `FileService.upload` posts `FormData` **without** manual `Content-Type`; returns the stored id/uuid.
- [ ] `<app-file-upload>` shows a preview via `/files/:uuid/low` and revokes object URLs (no leak).
- [ ] `<app-secure-image>` renders private images as blobs; placeholder when null.
- [ ] Dark theme + red accent + focus ring; IND/ENG keys match; "Using design-taste-frontend" announced.
