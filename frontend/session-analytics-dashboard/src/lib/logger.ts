type LogPrimitive = string | number | boolean | null | undefined;

export type LogContext = Record<string, LogPrimitive>;

export interface Logger {
  info(event: string, context?: LogContext): void;
  warn(event: string, context?: LogContext): void;
  error(event: string, context?: LogContext): void;
}

class ConsoleLogger implements Logger {
  info(event: string, context: LogContext = {}) {
    this.emit("info", event, context);
  }

  warn(event: string, context: LogContext = {}) {
    this.emit("warn", event, context);
  }

  error(event: string, context: LogContext = {}) {
    this.emit("error", event, context);
  }

  private emit(level: "info" | "warn" | "error", event: string, context: LogContext) {
    const payload = {
      level,
      event,
      timestamp: new Date().toISOString(),
      ...context
    };

    if (level === "error") {
      console.error(payload);
      return;
    }

    if (level === "warn") {
      console.warn(payload);
      return;
    }

    console.info(payload);
  }
}

export const logger: Logger = new ConsoleLogger();
