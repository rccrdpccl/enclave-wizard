export class Container {
  private readonly bindings = new Map<symbol, unknown>();

  register<T>(key: symbol, instance: T): void {
    this.bindings.set(key, instance);
  }

  resolve<T>(key: symbol): T {
    const instance = this.bindings.get(key);
    if (instance === undefined) {
      throw new Error(
        `No binding found for ${key.toString()}. Did you forget to register it?`,
      );
    }
    return instance as T;
  }
}
