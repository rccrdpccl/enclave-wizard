export declare class Container {
    private readonly bindings;
    register<T>(key: symbol, instance: T): void;
    resolve<T>(key: symbol): T;
}
//# sourceMappingURL=Container.d.ts.map