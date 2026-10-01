---
description: Build product-focused context for the TODO application
allowed-tools: Bash, Read, Glob, Grep
---

# Prepara el contexto desde el producto

Carga el contexto inicial de la aplicación desde la perspectiva de producto. No valores ni definas un cambio concreto: reconstruye qué problema resuelve hoy el producto, para quién y cómo se comporta. No modifiques ningún archivo.

## Exploración

1. Ejecuta `git status --short --branch` y arranca por las pantallas, rutas y tests de comportamiento disponibles.
2. Reconstruye los flujos de usuario observables: qué puede ver una persona, qué acciones puede realizar y qué resultado espera tras cada acción.
3. Lee las acciones, modelos, migraciones y templates que sostienen esos flujos para identificar reglas de dominio, estados y restricciones ya codificadas.
4. Revisa fixtures, textos de interfaz y pruebas para separar comportamiento confirmado de supuestos.
5. Identifica efectos secundarios, estados vacíos, errores y límites del producto que puedan verse afectados por un cambio.

## Resultado

Resume de forma concisa:

- propósito y comportamiento actual del producto;
- flujos de usuario y reglas de dominio observadas;
- criterios de aceptación implícitos en interfaz y tests;
- decisiones de producto que seguirían requiriendo confirmación humana;
- preguntas abiertas y comportamientos que un encargo futuro debería preservar o aclarar.

No inventes personas usuarias, prioridades ni requisitos. Distingue hechos observados de inferencias. El resultado es contexto inicial, no una especificación ni una propuesta de implementación.
