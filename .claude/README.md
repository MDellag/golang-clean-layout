# Claude Code Configuration

Este directorio contiene configuración específica para Claude Code (claude.ai/code).

## Estructura

```
.claude/
├── README.md                    # Este archivo
└── skills/                      # Custom skills
    └── clean-arch/              # Clean Architecture enforcer skill
        ├── SKILL.md             # Skill principal (ejecutable)
        ├── reference.md         # Referencia rápida y templates
        └── examples.md          # Ejemplos prácticos
```

## Skills Disponibles

### `/clean-arch` - Clean Architecture Enforcer

**Descripción:** Enforcer estricto de Clean Architecture y mejores prácticas de Go.

**Uso:**
```
/clean-arch
/clean-arch voy a crear un sistema de pagos
/clean-arch dónde va este código?
/clean-arch necesito integrar Stripe
```

**Cuándo usar:**
- ✅ Crear nuevos archivos o features
- ✅ Refactorizar código existente
- ✅ Dudas sobre estructura
- ✅ Planear implementaciones
- ✅ Revisar código

**Qué hace:**
- Guía sobre ubicación correcta de archivos
- Previene violaciones arquitectónicas
- Proporciona templates de código
- Valida flujo de dependencias
- Sugiere mejores prácticas

## Cómo funciona

Los skills son automáticamente descubiertos por Claude Code cuando trabajas en el proyecto. Puedes:

1. **Invocación manual:** Escribe `/clean-arch` en tu conversación
2. **Invocación automática:** Claude puede usar el skill cuando es relevante
3. **Consulta directa:** Lee los archivos `.md` directamente

## Documentación Adicional

Además de los skills, hay documentación de referencia en:

- `.claude-code/skills/` - Documentación detallada legacy
- `CLAUDE.md` - Overview del proyecto
- `CONTRIBUTING.md` - Guía de contribución

## Validación

Después de implementar cambios, valida con:

```bash
./scripts/validate_architecture.sh
```

## Mantenimiento

Para modificar el skill:

1. Edita `.claude/skills/clean-arch/SKILL.md`
2. Actualiza archivos de soporte si necesario
3. Prueba con `/clean-arch` en Claude Code
4. Documenta cambios

## Notas

- Los skills son específicos de este proyecto (scope: local)
- Para skills personales globales, usar `~/.claude/skills/`
- El archivo `SKILL.md` requiere frontmatter YAML
- Skills se cargan automáticamente, no requieren reinicio

---

**Última actualización:** 2026-01-22
