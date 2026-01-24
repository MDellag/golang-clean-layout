#!/bin/bash

# Clean Architecture Validator
# Este script verifica que el proyecto siga las reglas de Clean Architecture

set -e

ERRORS=0

echo "🔍 Validando Clean Architecture..."
echo ""

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
    ERRORS=$((ERRORS + 1))
}

warning() {
    echo -e "${YELLOW}⚠️  WARNING: $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# 1. Verificar que domain/entity no tenga tags de DB
echo "1. Verificando que entities no tengan tags de DB..."
if grep -r "gorm:\|bson:\|db:" internal/domain/entity/ 2>/dev/null; then
    error "Entities tienen tags de base de datos. Mueve estas a repositories/{db}/models/"
else
    success "Entities están limpias (sin tags de DB)"
fi
echo ""

# 2. Verificar que domain no importe infrastructure
echo "2. Verificando que domain no dependa de infrastructure..."
DOMAIN_IMPORTS=$(go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/domain/... 2>/dev/null | grep infrastructure || true)
if [ -n "$DOMAIN_IMPORTS" ]; then
    error "Domain layer importa infrastructure:\n$DOMAIN_IMPORTS"
else
    success "Domain layer es independiente"
fi
echo ""

# 3. Verificar que domain no importe delivery
echo "3. Verificando que domain no dependa de delivery..."
DOMAIN_DELIVERY=$(go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/domain/... 2>/dev/null | grep delivery || true)
if [ -n "$DOMAIN_DELIVERY" ]; then
    error "Domain layer importa delivery:\n$DOMAIN_DELIVERY"
else
    success "Domain no depende de delivery"
fi
echo ""

# 4. Verificar que domain no importe repositories
echo "4. Verificando que domain no dependa de repositories..."
DOMAIN_REPOS=$(go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/domain/... 2>/dev/null | grep repositories || true)
if [ -n "$DOMAIN_REPOS" ]; then
    error "Domain layer importa repositories:\n$DOMAIN_REPOS"
else
    success "Domain no depende de repositories"
fi
echo ""

# 5. Verificar que handlers no importen repositories directamente
echo "5. Verificando que handlers no accedan a repositories directamente..."
HANDLER_REPOS=$(go list -f '{{.ImportPath}} {{join .Imports " "}}' ./internal/delivery/handlers/... 2>/dev/null | grep repositories || true)
if [ -n "$HANDLER_REPOS" ]; then
    error "Handlers importan repositories directamente. Deben usar services:\n$HANDLER_REPOS"
else
    success "Handlers no acceden a repositories directamente"
fi
echo ""

# 6. Verificar estructura de carpetas requeridas
echo "6. Verificando estructura de carpetas..."
REQUIRED_DIRS=(
    "internal/domain/entity"
    "internal/domain/interfaces"
    "internal/domain/dto"
    "internal/services"
    "internal/repositories"
    "internal/delivery"
    "cmd"
    "config"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        success "$dir existe"
    else
        warning "$dir no existe (puede que aún no sea necesario)"
    fi
done
echo ""

# 7. Verificar que existan interfaces de repositories
echo "7. Verificando interfaces de repositories..."
if [ -f "internal/domain/interfaces/repositories.go" ]; then
    success "internal/domain/interfaces/repositories.go existe"
else
    warning "internal/domain/interfaces/repositories.go no existe"
fi
echo ""

# 8. Verificar que existan interfaces de services
echo "8. Verificando interfaces de services..."
if [ -f "internal/domain/interfaces/services.go" ]; then
    success "internal/domain/interfaces/services.go existe"
else
    warning "internal/domain/interfaces/services.go no existe"
fi
echo ""

# 9. Verificar que repositories tengan mappers
echo "9. Verificando estructura de repositories..."
for db_dir in internal/repositories/*/; do
    if [ -d "$db_dir" ]; then
        db_name=$(basename "$db_dir")
        if [ "$db_name" != "memory" ]; then
            if [ -d "${db_dir}models" ]; then
                success "$db_name tiene carpeta models/"
            else
                warning "$db_name debería tener carpeta models/"
            fi

            if [ -d "${db_dir}mappers" ]; then
                success "$db_name tiene carpeta mappers/"
            else
                warning "$db_name debería tener carpeta mappers/"
            fi
        fi
    fi
done
echo ""

# 10. Verificar dependencias circulares
echo "10. Verificando dependencias circulares..."
CIRCULAR=$(go list -f '{{.ImportPath}}' ./... 2>&1 | grep "import cycle" || true)
if [ -n "$CIRCULAR" ]; then
    error "Dependencias circulares detectadas:\n$CIRCULAR"
else
    success "No hay dependencias circulares"
fi
echo ""

# 11. Verificar que entities tengan métodos de negocio
echo "11. Verificando que entities tengan métodos (no sean anémicas)..."
for entity_file in internal/domain/entity/*.go; do
    if [ -f "$entity_file" ]; then
        entity_name=$(basename "$entity_file" .go)
        # Buscar métodos con receiver
        methods=$(grep -c "func ([a-z].*$entity_name)" "$entity_file" 2>/dev/null || echo "0")
        if [ "$methods" -gt 0 ]; then
            success "$entity_name tiene $methods métodos de negocio"
        else
            warning "$entity_name podría ser anémica (no tiene métodos con receiver)"
        fi
    fi
done
echo ""

# 12. Verificar que DTOs tengan tags JSON
echo "12. Verificando que DTOs tengan tags de serialización..."
if [ -d "internal/domain/dto" ]; then
    DTO_COUNT=$(find internal/domain/dto -name "*.go" -type f | wc -l)
    DTO_WITH_TAGS=$(grep -r "json:" internal/domain/dto/*.go 2>/dev/null | wc -l || echo "0")

    if [ "$DTO_COUNT" -gt 0 ] && [ "$DTO_WITH_TAGS" -eq 0 ]; then
        warning "DTOs existen pero ninguno tiene tags JSON"
    elif [ "$DTO_WITH_TAGS" -gt 0 ]; then
        success "DTOs tienen tags de serialización"
    fi
fi
echo ""

# 13. Tests
echo "13. Verificando existencia de tests..."
TEST_COUNT=$(find . -name "*_test.go" -type f | wc -l)
if [ "$TEST_COUNT" -gt 0 ]; then
    success "Proyecto tiene $TEST_COUNT archivos de test"
else
    warning "No se encontraron tests"
fi
echo ""

# Resumen
echo "================================"
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ Validación completada sin errores!${NC}"
    echo ""
    echo "El proyecto sigue correctamente las reglas de Clean Architecture."
    exit 0
else
    echo -e "${RED}❌ Validación completada con $ERRORS errores${NC}"
    echo ""
    echo "Por favor, corrige los errores antes de continuar."
    echo "Consulta .claude-code/skills/clean-arch.md para más detalles."
    exit 1
fi
