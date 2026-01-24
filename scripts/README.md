# Scripts de Utilidad

## validate_architecture.sh

Script de validación de Clean Architecture que verifica el cumplimiento de las reglas arquitectónicas del proyecto.

### Uso

```bash
./scripts/validate_architecture.sh
```

### Qué valida

1. **Pureza del Domain Layer:**
   - Entities sin tags de DB
   - Domain no importa infrastructure
   - Domain no importa delivery
   - Domain no importa repositories

2. **Separación de Concerns:**
   - Handlers no acceden a repositories directamente
   - Uso de interfaces para dependencies

3. **Estructura de Carpetas:**
   - Existencia de carpetas requeridas
   - Estructura correcta de repositories (models/ y mappers/)

4. **Calidad del Código:**
   - Entities no anémicas (con métodos de negocio)
   - DTOs con tags de serialización
   - Ausencia de dependencias circulares

5. **Testing:**
   - Existencia de tests

### Interpretación de Resultados

**✅ Success (Verde):** La verificación pasó correctamente.

**⚠️ Warning (Amarillo):** Advertencia que puede indicar una mejora posible, pero no es crítica.

**❌ Error (Rojo):** Violación de reglas de Clean Architecture que debe corregirse.

### Exit Codes

- `0`: Validación exitosa (sin errores)
- `1`: Validación falló (con errores)

### Integración con CI/CD

Puedes integrar este script en tu pipeline de CI/CD:

```yaml
# .github/workflows/validate.yml
- name: Validate Architecture
  run: ./scripts/validate_architecture.sh
```

### Ejemplos de Errores Comunes

#### Error: Entities con tags de DB

```
❌ ERROR: Entities tienen tags de base de datos
```

**Solución:**
- Mover tags a `internal/repositories/{db}/models/`
- Usar mappers para convertir models ↔ entities

#### Error: Domain importa infrastructure

```
❌ ERROR: Domain layer importa infrastructure
```

**Solución:**
- Definir interfaces en `internal/domain/interfaces/`
- Implementar en infrastructure
- Inyectar dependencias

#### Error: Handlers importan repositories

```
❌ ERROR: Handlers importan repositories directamente
```

**Solución:**
- Handlers solo deben usar services
- Services orquestan repositories

### Desarrollo del Script

Para agregar nuevas validaciones:

1. Crea una nueva función de validación
2. Llámala en el flujo principal
3. Usa `error()`, `warning()`, o `success()` para reportar
4. Incrementa `ERRORS` solo para violaciones críticas

### Notas

- El script usa `set -e` (exit on error) para comandos críticos
- Las warnings no causan que el script falle
- Los errores sí causan que el script retorne exit code 1
