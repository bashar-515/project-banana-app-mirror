export interface View<Type = void> {
  init(arg: Type): void;
  tearDown(): void;
}
