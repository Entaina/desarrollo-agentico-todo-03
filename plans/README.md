# Planes de implementación

Este directorio es la ubicación estable de los planes generados por las skills `/bug`, `/chore` y `/feature`, definidas en `.claude/skills/`.

Cada plan concreta un encargo después de investigar el estado real del repositorio. Debe ser autocontenido y conservar el alcance, la evidencia, los archivos relevantes, los pasos, la validación y los riesgos necesarios para revisarlo sin depender de la conversación de planificación.

Una vez aprobado, el plan se entrega a `/implement plans/<plan>.md`, que ejecuta sus decisiones sin rediseñar ni ampliar la tarea.

`/dev-flow <requisito>` clasifica el requisito como bug, chore o feature, encadena la skill de planificación correspondiente con `/implement` y añade la validación del CI y el despliegue local con prueba de humo.

Las invocaciones conservan sus nombres. Si ya tenías una sesión de Claude abierta antes de crear `.claude/skills/`, ejecuta `/reload-skills` o inicia una nueva sesión para que descubra las skills.

El [trigger de issues](../adws/README.md) puede invocar el workflow que construyas con estas piezas; no sustituye la planificación ni la implementación.
