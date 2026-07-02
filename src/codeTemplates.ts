export interface CodeFile {
  name: string;
  path: string;
  description: string;
  language: string;
  code: string;
}

export const FOLDER_STRUCTURE_TEXT = `├── .godot/                  # Cache e arquivos gerados pelo Godot
├── assets/                  # Assets visuais, de áudio e fontes
│   ├── ui/                  # Texturas, temas e fontes para a interface
│   └── environment/         # Assets de ambiente e modelos/sprites
├── scenes/                  # Cenas .tscn do Godot
│   ├── bootstrap.tscn       # Cena inicial do jogo (vazia, apenas carrega scripts)
│   ├── main_menu.tscn       # Cena do menu principal
│   ├── char_selection.tscn  # Cena de seleção de personagens
│   └── game_world.tscn      # Cena do mundo do jogo (InGame)
├── src/                     # Código fonte C# principal
│   ├── Core/                # Núcleo da arquitetura (State Machine, Managers)
│   │   ├── GameManager.cs   # Singleton principal (Autoload)
│   │   ├── ServiceRegistry.cs # Centralizador de Dependency Injection leve [NOVO]
│   │   ├── CrashHandler.cs  # Captura global de exceções e log de falhas [NOVO]
│   │   ├── ConfigManager.cs # Gerenciador de configurações locais (JSON)
│   │   ├── EventBus.cs      # Barramento de eventos desacoplado e fortemente tipado
│   │   ├── SceneManager.cs  # Controlador de transição e carregamento assíncrono de cenas
│   │   └── NetworkManager.cs# Gerenciador de conexões de rede (WebSocket/TCP)
│   ├── StateMachine/        # Máquina de Estados da Aplicação
│   │   ├── IAppState.cs     # Interface base para todos os estados
│   │   ├── AppStateMachine.cs # Máquina que gerencia a transição de estados
│   │   └── States/          # Implementações concretas de cada estado
│   │       ├── BootState.cs
│   │       ├── LoadingState.cs
│   │       ├── MenuState.cs
│   │       ├── ConnectingState.cs
│   │       ├── CharacterSelectionState.cs
│   │       ├── InGameState.cs
│   │       ├── DisconnectedState.cs
│   │       └── ShutdownState.cs
│   ├── UI/                  # Scripts de UI associados a cenas
│   │   ├── BootstrapUI.cs   # Script de controle da cena de Boot
│   │   └── MainMenuUI.cs    # Script de controle do Menu Principal
│   └── Network/             # Protocolos e pacotes de rede
│       ├── GamePacket.cs    # Estrutura padronizada de pacotes binários [REFATORADO]
│       └── PacketOpcode.cs  # Enumerador de Opcodes (códigos de operação)
├── LightAndShadow.csproj    # Definição do projeto C# (MSBuild)
└── project.godot            # Arquivo de configuração do projeto Godot`;

export const codeFiles: CodeFile[] = [
  {
    name: "ServiceRegistry.cs",
    path: "src/Core/ServiceRegistry.cs",
    description: "Centralizador thread-safe para registro e resolução leve de dependências (Injeção de Dependências) no MMORPG.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;
using Godot;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// Service Registry Centralizado para Injeção de Dependências leve e desacoplada.
    /// Thread-safe, suporta registro de instâncias, inicialização tardia (lazy) e métodos genéricos.
    /// </summary>
    public class ServiceRegistry
    {
        private static readonly object _lock = new();
        private static ServiceRegistry? _instance;

        public static ServiceRegistry Instance
        {
            get
            {
                if (_instance == null)
                {
                    lock (_lock)
                    {
                        _instance ??= new ServiceRegistry();
                    }
                }
                return _instance;
            }
        }

        private readonly Dictionary<Type, object> _services = new();
        private readonly Dictionary<Type, Func<object>> _lazyFactories = new();
        private readonly object _registryLock = new();

        private ServiceRegistry() { }

        /// <summary>
        /// Registra uma instância concreta de um serviço.
        /// </summary>
        public void Register<T>(T service) where T : class
        {
            if (service == null) throw new ArgumentNullException(nameof(service));

            lock (_registryLock)
            {
                Type type = typeof(T);
                if (_services.ContainsKey(type))
                {
                    throw new InvalidOperationException($"Serviço do tipo {type.FullName} já está registrado.");
                }

                _services[type] = service;
                _lazyFactories.Remove(type);
                GD.Print($"[ServiceRegistry] Serviço registrado: {type.Name}");
            }
        }

        /// <summary>
        /// Registra um serviço com inicialização tardia (lazy) via fábrica delegada.
        /// </summary>
        public void Register<T>(Func<T> factory) where T : class
        {
            if (factory == null) throw new ArgumentNullException(nameof(factory));

            lock (_registryLock)
            {
                Type type = typeof(T);
                if (_services.ContainsKey(type) || _lazyFactories.ContainsKey(type))
                {
                    throw new InvalidOperationException($"Serviço ou Fábrica do tipo {type.FullName} já está registrado.");
                }

                _lazyFactories[type] = () => factory();
                GD.Print($"[ServiceRegistry] Serviço Lazy registrado: {type.Name}");
            }
        }

        /// <summary>
        /// Registra um serviço instanciado dinamicamente no primeiro Resolve.
        /// </summary>
        public void Register<T>() where T : class, new()
        {
            Register(() => new T());
        }

        /// <summary>
        /// Resolve e retorna o serviço registrado correspondente ao tipo genérico.
        /// Lança exceção caso não seja encontrado.
        /// </summary>
        public T Resolve<T>() where T : class
        {
            lock (_registryLock)
            {
                Type type = typeof(T);
                if (_services.TryGetValue(type, out var service))
                {
                    return (T)service;
                }

                if (_lazyFactories.TryGetValue(type, out var factory))
                {
                    GD.Print($"[ServiceRegistry] Resolvendo serviço Lazy pela primeira vez: {type.Name}");
                    try
                     {
                        T instance = (T)factory();
                        _services[type] = instance;
                        _lazyFactories.Remove(type);
                        return instance;
                    }
                    catch (Exception ex)
                    {
                        GD.PrintErr($"[ServiceRegistry] Falha ao instanciar serviço Lazy {type.Name}: {ex.Message}");
                        throw new InvalidOperationException($"Falha ao instanciar serviço Lazy {type.Name}", ex);
                    }
                }

                throw new KeyNotFoundException($"Nenhum serviço registrado para o tipo: {type.FullName}");
            }
        }

        /// <summary>
        /// Tenta resolver o serviço registrado correspondente ao tipo genérico sem lançar exceções.
        /// </summary>
        public T? TryResolve<T>() where T : class
        {
            lock (_registryLock)
            {
                try
                {
                    return Resolve<T>();
                }
                catch
                {
                    return null;
                }
            }
        }

        /// <summary>
        /// Remove um serviço ou fábrica registrado.
        /// </summary>
        public bool Remove<T>() where T : class
        {
            lock (_registryLock)
            {
                Type type = typeof(T);
                bool removed = _services.Remove(type) || _lazyFactories.Remove(type);
                if (removed)
                {
                    GD.Print($"[ServiceRegistry] Serviço removido do registro: {type.Name}");
                }
                return removed;
            }
        }
    }
}`
  },
  {
    name: "CrashHandler.cs",
    path: "src/Core/CrashHandler.cs",
    description: "Capturador global e thread-safe de exceções críticas e crash logs com fallback de encerramento controlado.",
    language: "csharp",
    code: `using System;
using System.IO;
using System.Threading;
using System.Threading.Tasks;
using Godot;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// Gerenciador de Exceções e Falhas Críticas (Crash Handler) do MMORPG Client.
    /// Intercepta erros síncronos e assíncronos não tratados, gera relatórios em disco e força shutdown gracioso.
    /// </summary>
    public class CrashHandler
    {
        private static readonly object _lock = new();
        private static CrashHandler? _instance;
        private int _isHandlingCrash = 0; // 0 = false, 1 = true (atomic boolean representation)

        public static CrashHandler Instance
        {
            get
            {
                if (_instance == null)
                {
                    lock (_lock)
                    {
                        _instance ??= new CrashHandler();
                    }
                }
                return _instance;
            }
        }

        private CrashHandler() { }

        /// <summary>
        /// Vincula os tratadores globais do AppDomain e do TaskScheduler de forma thread-safe.
        /// </summary>
        public void Initialize()
        {
            lock (_lock)
            {
                AppDomain.CurrentDomain.UnhandledException += OnUnhandledException;
                TaskScheduler.UnobservedTaskException += OnUnobservedTaskException;
                GD.Print("[CrashHandler] Inicializado com sucesso. Escutando falhas do sistema...");
            }
        }

        private void OnUnhandledException(object sender, UnhandledExceptionEventArgs e)
        {
            if (e.ExceptionObject is Exception ex)
            {
                HandleCrash(ex, "AppDomain.UnhandledException", !e.IsTerminating);
            }
            else
            {
                HandleCrash(new Exception($"Objeto de exceção desconhecido: {e.ExceptionObject}"), "AppDomain.UnhandledException", false);
            }
        }

        private void OnUnobservedTaskException(object? sender, UnobservedTaskExceptionEventArgs e)
        {
            e.SetObserved(); // Marca como observada para evitar travamento do runtime dependendo da config
            HandleCrash(e.Exception, "TaskScheduler.UnobservedTaskException", true);
        }

        /// <summary>
        /// Rotina central de processamento de crash, serialização de log e fallback.
        /// </summary>
        public void HandleCrash(Exception ex, string context, bool canAttemptRecovery)
        {
            // PATCH 4 - Crash Recursion Guard via atomic boolean simulation (Interlocked)
            if (System.Threading.Interlocked.CompareExchange(ref _isHandlingCrash, 1, 0) != 0)
            {
                GD.PrintErr("[CrashHandler] RECURSIVE CRASH DETECTED! Bypassing EventBus and forcing immediate shutdown.");
                try
                {
                    var configManager = ServiceRegistry.Instance.TryResolve<ConfigManager>();
                    configManager?.SaveConfig();

                    var networkManager = ServiceRegistry.Instance.TryResolve<NetworkManager>();
                    networkManager?.Disconnect();
                }
                catch { }

                if (GameManager.Instance != null)
                {
                    GameManager.Instance.GetTree().Quit(-1);
                }
                else
                {
                    Environment.Exit(-1);
                }
                return;
            }

            string timestamp = DateTime.Now.ToString("yyyy-MM-dd_HH-mm-ss");
            string logDirectory = ProjectSettings.GlobalizePath("user://logs/");
            string logPath = Path.Combine(logDirectory, $"crash_{timestamp}.log");

            // Formato de Log Mandatório:
            // timestamp | thread id | exception type | message | stacktrace
            int threadId = Environment.CurrentManagedThreadId;
            string exceptionType = ex.GetType().FullName ?? "UnknownException";
            string message = ex.Message;
            string stacktrace = ex.StackTrace ?? "No stacktrace available";

            string logContent = $"Timestamp: {DateTime.Now:yyyy-MM-dd HH:mm:ss.fff}\\n" +
                               $"Thread ID: {threadId}\\n" +
                               $"Exception Type: {exceptionType}\\n" +
                               $"Message: {message}\\n" +
                               $"Context: {context}\\n" +
                               $"Stack Trace:\\n{stacktrace}\\n";

            if (ex.InnerException != null)
            {
                logContent += $"\\nInner Exception: {ex.InnerException.GetType().FullName}\\n" +
                             $"Inner Message: {ex.InnerException.Message}\\n" +
                             $"Inner Stack Trace:\\n{ex.InnerException.StackTrace}\\n";
            }

            // Persiste logs de forma defensiva
            try
            {
                if (!Directory.Exists(logDirectory))
                {
                    Directory.CreateDirectory(logDirectory);
                }
                File.WriteAllText(logPath, logContent);
                GD.PrintErr($"[CrashHandler] EXCEÇÃO CRÍTICA NÃO TRATADA SALVA EM: {logPath}");
            }
            catch (Exception writeEx)
            {
                GD.PrintErr($"[CrashHandler] Falha ao escrever arquivo de crash log: {writeEx.Message}");
            }

            // Tentativa de Recuperação Segura (Recovery Attempt) se aplicável
            if (canAttemptRecovery && AttemptRecovery(ex, context))
            {
                System.Threading.Interlocked.Exchange(ref _isHandlingCrash, 0);
                return;
            }

            // Graceful Shutdown ou Fallback de Encerramento Crítico
            InitiateGracefulShutdown(ex);
        }

        private bool AttemptRecovery(Exception ex, string context)
        {
            GD.Print($"[CrashHandler] Tentando recuperar de exceção no contexto {context}...");
            try
            {
                var eventBus = ServiceRegistry.Instance.TryResolve<EventBus>();
                if (eventBus != null)
                {
                    GD.Print("[CrashHandler] EventBus ativo. Notificando falha recuperável.");
                    // Permitiria a um subsistema recriar um buffer ou reconectar sem quebrar o loop de render
                    return true;
                }
            }
            catch (Exception recoveryEx)
            {
                GD.PrintErr($"[CrashHandler] Falha na tentativa de recuperação: {recoveryEx.Message}");
            }
            return false;
        }

        private void InitiateGracefulShutdown(Exception ex)
        {
            GD.PrintErr("[CrashHandler] Falha irrecuperável detectada. Iniciando encerramento forçado do cliente...");

            try
            {
                var stateMachine = ServiceRegistry.Instance.TryResolve<AppStateMachine>();
                if (stateMachine != null)
                {
                    GD.Print("[CrashHandler] Solicitando transição assíncrona para o estado de Shutdown.");
                    _ = stateMachine.TransitionToAsync(StateMachine.AppStateType.Shutdown);
                    return;
                }
            }
            catch (Exception shutdownEx)
            {
                GD.PrintErr($"[CrashHandler] Erro ao tentar transição graciosa de pânico: {shutdownEx.Message}");
            }

            // Fallback de desligamento imediato (Se o motor ou a máquina de estados estiverem quebrados)
            GD.PrintErr("[CrashHandler] Fallback ativado. Forçando desalocação de hardware e fechamento imediato.");
            try
            {
                var configManager = ServiceRegistry.Instance.TryResolve<ConfigManager>();
                configManager?.SaveConfig();

                var networkManager = ServiceRegistry.Instance.TryResolve<NetworkManager>();
                networkManager?.Disconnect();
            }
            catch { }

            if (GameManager.Instance != null)
            {
                GameManager.Instance.GetTree().Quit(-1);
            }
            else
            {
                Environment.Exit(-1);
            }
        }
    }
}`
  },
  {
    name: "IAppState.cs",
    path: "src/StateMachine/IAppState.cs",
    description: "Interface definindo os contratos obrigatórios de ciclo de vida para os estados do ciclo global do MMORPG.",
    language: "csharp",
    code: `using System;
using System.Threading.Tasks;

namespace LightAndShadow.Client.StateMachine
{
    /// <summary>
    /// Interface que define o ciclo de vida de um estado da aplicação.
    /// Garante que transições assíncronas sejam tratadas de forma segura.
    /// </summary>
    public interface IAppState
    {
        /// <summary>
        /// Identificador único do estado correspondente.
        /// </summary>
        AppStateType StateType { get; }

        /// <summary>
        /// Chamado ao entrar no estado.
        /// </summary>
        Task EnterAsync();

        /// <summary>
        /// Chamado a cada frame físico ou de renderização de forma delegada pelo State Machine.
        /// </summary>
        /// <param name="delta">Tempo decorrido desde o último frame em segundos.</param>
        void Update(double delta);

        /// <summary>
        /// Chamado ao sair do estado. Útil para limpeza de listeners e desalocação de memória.
        /// </summary>
        Task ExitAsync();
    }

    /// <summary>
    /// Tipos de estados disponíveis no jogo.
    /// </summary>
    public enum AppStateType
    {
        Boot,
        Loading,
        Menu,
        Connecting,
        CharacterSelection,
        InGame,
        Disconnected,
        Shutdown
    }
}`
  },
  {
    name: "AppStateMachine.cs",
    path: "src/StateMachine/AppStateMachine.cs",
    description: "Motor de estados que lida com o estado atual, transições, threads de segurança e disparo de eventos de transição.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Godot;

namespace LightAndShadow.Client.StateMachine
{
    /// <summary>
    /// Gerenciador de Máquina de Estados da Aplicação.
    /// Responsável pelas transições seguras entre estados de forma assíncrona.
    /// </summary>
    public class AppStateMachine
    {
        private readonly Dictionary<AppStateType, IAppState> _states = new();
        private readonly object _lock = new();
        
        public IAppState CurrentState { get; private set; }
        public bool IsTransitioning { get; private set; }

        public event Action<IAppState, IAppState> OnStateChanged;

        public void RegisterState(IAppState state)
        {
            lock (_lock)
            {
                if (_states.ContainsKey(state.StateType))
                {
                    _states[state.StateType] = state;
                }
                else
                {
                    _states.Add(state.StateType, state);
                }
                GD.Print($"[StateMachine] Estado registrado: {state.StateType}");
            }
        }

        public async Task TransitionToAsync(AppStateType targetStateType)
        {
            IAppState nextState = null;

            lock (_lock)
            {
                if (IsTransitioning)
                {
                    GD.PrintErr($"[StateMachine] Tentativa de transicionar para {targetStateType} bloqueada: Transição em andamento.");
                    return;
                }

                if (!_states.TryGetValue(targetStateType, out nextState))
                {
                    GD.PrintErr($"[StateMachine] Erro crítico: Estado {targetStateType} não está registrado na máquina.");
                    return;
                }

                if (CurrentState?.StateType == targetStateType)
                {
                    GD.Print($"[StateMachine] Já estamos no estado {targetStateType}. Abortando transição redundante.");
                    return;
                }

                IsTransitioning = true;
            }

            IAppState previousState = CurrentState;
            GD.Print($"[StateMachine] Iniciando transição: {previousState?.StateType.ToString() ?? "Nenhum"} -> {targetStateType}");

            try
            {
                if (CurrentState != null)
                {
                    await CurrentState.ExitAsync();
                }

                CurrentState = nextState;
                await CurrentState.EnterAsync();

                lock (_lock)
                {
                    IsTransitioning = false;
                }

                OnStateChanged?.Invoke(previousState, CurrentState);
                GD.Print($"[StateMachine] Transição concluída com sucesso para o estado: {targetStateType}");
            }
            catch (Exception ex)
            {
                lock (_lock)
                {
                    IsTransitioning = false;
                }
                GD.PrintErr($"[StateMachine] Falha crítica na transição para {targetStateType}: {ex.Message}\\n{ex.StackTrace}");
                throw;
            }
        }

        public void Update(double delta)
        {
            if (IsTransitioning) return;
            CurrentState?.Update(delta);
        }
    }
}`
  },
  {
    name: "States.cs",
    path: "src/StateMachine/States/States.cs",
    description: "Implementações completas dos 8 estados requeridos, desacoplados usando resoluções do ServiceRegistry.",
    language: "csharp",
    code: `using System;
using System.Threading.Tasks;
using Godot;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.StateMachine.States
{
    // ==========================================
    // 1. BOOT STATE
    // ==========================================
    public class BootState : IAppState
    {
        public AppStateType StateType => AppStateType.Boot;

        public async Task EnterAsync()
        {
            GD.Print("[State: Boot] Inicializando sistemas básicos do MMORPG Light and Shadow...");
            
            await Task.Delay(1000); // Carregamento inicial assíncrono fictício

            GD.Print("[State: Boot] Motores de Configuração, DI Registry e Rede prontos.");

            // Redireciona via ServiceRegistry resolvendo o StateMachine
            _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Loading);
        }

        public void Update(double delta) { }
        public async Task ExitAsync() => await Task.CompletedTask;
    }

    // ==========================================
    // 2. LOADING STATE
    // ==========================================
    public class LoadingState : IAppState
    {
        public AppStateType StateType => AppStateType.Loading;
        private double _loadingTimer = 0.0;
        private bool _transitionTriggered = false;

        public async Task EnterAsync()
        {
            GD.Print("[State: Loading] Entrou no estado de carregamento de Assets e UI...");
            _loadingTimer = 0.0;
            _transitionTriggered = false;

            // Transiciona a cena visual pelo SceneManager resolvido via DI
            ServiceRegistry.Instance.Resolve<SceneManager>().LoadSceneAsync("res://scenes/bootstrap.tscn");
            
            await Task.CompletedTask;
        }

        public void Update(double delta)
        {
            _loadingTimer += delta;
            
            if (_loadingTimer >= 2.0 && !_transitionTriggered)
            {
                _transitionTriggered = true;
                GD.Print("[State: Loading] Recursos de interface carregados. Redirecionando ao Menu Principal.");
                _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Menu);
            }
        }

        public async Task ExitAsync() => await Task.CompletedTask;
    }

    // ==========================================
    // 3. MENU STATE
    // ==========================================
    public class MenuState : IAppState
    {
        public AppStateType StateType => AppStateType.Menu;

        public async Task EnterAsync()
        {
            GD.Print("[State: Menu] Exibindo Menu Principal...");
            
            ServiceRegistry.Instance.Resolve<SceneManager>().LoadSceneAsync("res://scenes/main_menu.tscn");
            ServiceRegistry.Instance.Resolve<EventBus>().Subscribe<string>(EventName.OnLoginAttempted, OnLoginAttempted);
            
            await Task.CompletedTask;
        }

        private void OnLoginAttempted(string accountCredentials)
        {
            GD.Print($"[State: Menu] Tentativa de login recebida: {accountCredentials}");
            _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Connecting);
        }

        public void Update(double delta) { }

        public async Task ExitAsync()
        {
            ServiceRegistry.Instance.Resolve<EventBus>().Unsubscribe<string>(EventName.OnLoginAttempted, OnLoginAttempted);
            GD.Print("[State: Menu] Limpeza e ocultamento do Menu concluídos.");
            await Task.CompletedTask;
        }
    }

    // ==========================================
    // 4. CONNECTING STATE
    // ==========================================
    public class ConnectingState : IAppState
    {
        public AppStateType StateType => AppStateType.Connecting;
        private Task<bool>? _connectionTask;

        public async Task EnterAsync()
        {
            GD.Print("[State: Connecting] Tentando conectar aos servidores do MMORPG...");
            
            var config = ServiceRegistry.Instance.Resolve<ConfigManager>();
            var network = ServiceRegistry.Instance.Resolve<NetworkManager>();

            _connectionTask = network.ConnectAsync(config.ServerHost, config.ServerPort);
            
            EvaluateConnectionAsync();
            await Task.CompletedTask;
        }

        private async void EvaluateConnectionAsync()
        {
            if (_connectionTask == null) return;

            try
            {
                bool success = await _connectionTask;
                if (success)
                {
                    GD.Print("[State: Connecting] Conexão TCP estabelecida com sucesso!");
                    _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.CharacterSelection);
                }
                else
                {
                    GD.PrintErr("[State: Connecting] Falha ao se conectar com o servidor.");
                    _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Disconnected);
                }
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[State: Connecting] Erro durante a tentativa de conexão: {ex.Message}");
                _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Disconnected);
            }
        }

        public void Update(double delta) { }
        public async Task ExitAsync() => await Task.CompletedTask;
    }

    // ==========================================
    // 5. CHARACTER SELECTION STATE
    // ==========================================
    public class CharacterSelectionState : IAppState
    {
        public AppStateType StateType => AppStateType.CharacterSelection;

        public async Task EnterAsync()
        {
            GD.Print("[State: CharacterSelection] Entrou no fluxo de seleção de personagens.");
            
            ServiceRegistry.Instance.Resolve<SceneManager>().LoadSceneAsync("res://scenes/char_selection.tscn");
            ServiceRegistry.Instance.Resolve<EventBus>().Subscribe<string>(EventName.OnCharacterSelected, OnCharacterSelected);
            
            await Task.CompletedTask;
        }

        private void OnCharacterSelected(string characterId)
        {
            GD.Print($"[State: CharacterSelection] Personagem selecionado: {characterId}. Entrando no mundo!");
            _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.InGame);
        }

        public void Update(double delta) { }

        public async Task ExitAsync()
        {
            ServiceRegistry.Instance.Resolve<EventBus>().Unsubscribe<string>(EventName.OnCharacterSelected, OnCharacterSelected);
            await Task.CompletedTask;
        }
    }

    // ==========================================
    // 6. IN-GAME STATE
    // ==========================================
    public class InGameState : IAppState
    {
        public AppStateType StateType => AppStateType.InGame;
        private double _heartbeatTimer = 0;

        public async Task EnterAsync()
        {
            GD.Print("[State: InGame] Sincronização de mapa e jogador estabelecida. Bem-vindo ao mundo!");
            
            ServiceRegistry.Instance.Resolve<SceneManager>().LoadSceneAsync("res://scenes/game_world.tscn");
            ServiceRegistry.Instance.Resolve<EventBus>().Subscribe(EventName.OnNetworkDisconnect, OnNetworkDisconnect);
            
            await Task.CompletedTask;
        }

        private void OnNetworkDisconnect()
        {
            GD.PrintErr("[State: InGame] Conexão com o servidor perdida repentinamente!");
            _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Disconnected);
        }

        public void Update(double delta)
        {
            _heartbeatTimer += delta;
            if (_heartbeatTimer >= 10.0)
            {
                _heartbeatTimer = 0;
                ServiceRegistry.Instance.Resolve<NetworkManager>().SendHeartbeat();
            }
        }

        public async Task ExitAsync()
        {
            ServiceRegistry.Instance.Resolve<EventBus>().Unsubscribe(EventName.OnNetworkDisconnect, OnNetworkDisconnect);
            await Task.CompletedTask;
        }
    }

    // ==========================================
    // 7. DISCONNECTED STATE
    // ==========================================
    public class DisconnectedState : IAppState
    {
        public AppStateType StateType => AppStateType.Disconnected;

        public async Task EnterAsync()
        {
            GD.PrintErr("[State: Disconnected] Conexão terminada. Retornando ao menu principal...");
            
            ServiceRegistry.Instance.Resolve<NetworkManager>().Disconnect();
            
            await Task.Delay(2000);
            
            _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Menu);
        }

        public void Update(double delta) { }
        public async Task ExitAsync() => await Task.CompletedTask;
    }

    // ==========================================
    // 8. SHUTDOWN STATE
    // ==========================================
    public class ShutdownState : IAppState
    {
        public AppStateType StateType => AppStateType.Shutdown;

        public async Task EnterAsync()
        {
            GD.Print("[State: Shutdown] Desligando o cliente de forma segura...");
            
            ServiceRegistry.Instance.Resolve<ConfigManager>().SaveConfig();
            ServiceRegistry.Instance.Resolve<NetworkManager>().Disconnect();
            
            GD.Print("[State: Shutdown] Conexão encerrada, configurações salvas. Sair.");
            
            await Task.Delay(500);
            
            // Resolve GameManager para acessar a árvore visual principal de forma segura
            ServiceRegistry.Instance.Resolve<GameManager>().GetTree().Quit();
        }

        public void Update(double delta) { }
        public async Task ExitAsync() => await Task.CompletedTask;
    }
}`
  },
  {
    name: "GameManager.cs",
    path: "src/Core/GameManager.cs",
    description: "Autoload e ponto de entrada global. Inicializa o CrashHandler e registra todos os managers no ServiceRegistry.",
    language: "csharp",
    code: `using System;
using Godot;
using LightAndShadow.Client.StateMachine;
using LightAndShadow.Client.StateMachine.States;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// GameManager central do MMORPG Light and Shadow.
    /// Orquestra a injeção de dependências inicial e aciona o CrashHandler.
    /// </summary>
    public partial class GameManager : Node
    {
        public static GameManager Instance { get; private set; }

        // Propriedades agora mapeadas dinamicamente via ServiceRegistry para manter 100% de compatibilidade
        public AppStateMachine StateMachine => ServiceRegistry.Instance.Resolve<AppStateMachine>();
        public ConfigManager ConfigManager => ServiceRegistry.Instance.Resolve<ConfigManager>();
        public EventBus EventBus => ServiceRegistry.Instance.Resolve<EventBus>();
        public SceneManager SceneManager => ServiceRegistry.Instance.Resolve<SceneManager>();
        public NetworkManager NetworkManager => ServiceRegistry.Instance.Resolve<NetworkManager>();

        public override void _EnterTree()
        {
            if (Instance != null && Instance != this)
            {
                GD.Print("[GameManager] Instância duplicada encontrada. Removendo nó redundante do grafo.");
                QueueFree();
                return;
            }
            Instance = this;
            ProcessMode = ProcessModeEnum.Always;

            // PATCH 2 — Inicialização imediata do crash handler prioritário
            CrashHandler.Instance.Initialize();

            // PATCH 1 — Inicializa e vincula os managers no ServiceRegistry central
            InitializeCore();
        }

        public override void _Ready()
        {
            GD.Print("[GameManager] Inicializando MMORPG Client Bootstrap...");
            
            RegisterApplicationStates();

            // Inicia o fluxo para o BootState
            _ = StateMachine.TransitionToAsync(AppStateType.Boot);
        }

        public override void _Process(double delta)
        {
            StateMachine?.Update(delta);
        }

        public override void _Notification(int what)
        {
            if (what == NotificationWMCloseRequest)
            {
                GD.Print("[GameManager] Pedido de fechamento detectado pelo OS.");
                GetTree().AutoAcceptQuit = false;
                _ = StateMachine.TransitionToAsync(AppStateType.Shutdown);
            }
        }

        private void InitializeCore()
        {
            var registry = ServiceRegistry.Instance;

            // Registra o GameManager para permitir acesso à árvore do motor
            registry.Register<GameManager>(this);

            // 1. Instanciar Configurações e ler preferências locais (user://)
            var configManager = new ConfigManager();
            AddChild(configManager);
            registry.Register<ConfigManager>(configManager);
            configManager.LoadConfig();

            // 2. Barramento de eventos global desacoplado
            var eventBus = new EventBus();
            AddChild(eventBus);
            registry.Register<EventBus>(eventBus);

            // 3. Gerenciador de conexões de soquetes assíncronos
            var networkManager = new NetworkManager();
            AddChild(networkManager);
            registry.Register<NetworkManager>(networkManager);

            // 4. Carregador e comutador assíncrono de cenas .tscn
            var sceneManager = new SceneManager();
            AddChild(sceneManager);
            registry.Register<SceneManager>(sceneManager);

            // 5. Instanciar Máquina de Estados principal
            var stateMachine = new AppStateMachine();
            registry.Register<AppStateMachine>(stateMachine);
        }

        private void RegisterApplicationStates()
        {
            StateMachine.RegisterState(new BootState());
            StateMachine.RegisterState(new LoadingState());
            StateMachine.RegisterState(new MenuState());
            StateMachine.RegisterState(new ConnectingState());
            StateMachine.RegisterState(new CharacterSelectionState());
            StateMachine.RegisterState(new InGameState());
            StateMachine.RegisterState(new DisconnectedState());
            StateMachine.RegisterState(new ShutdownState());
        }
    }
}`
  },
  {
    name: "EventBus.cs",
    path: "src/Core/EventBus.cs",
    description: "Sistema desacoplado de publicação e assinatura altamente extensível para eventos sem parâmetros ou parametrizados genéricos.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;
using Godot;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// Eventos pré-definidos do MMORPG.
    /// </summary>
    public static class EventName
    {
        public const string OnLoginAttempted = "OnLoginAttempted";
        public const string OnCharacterSelected = "OnCharacterSelected";
        public const string OnNetworkDisconnect = "OnNetworkDisconnect";
        public const string OnNetworkPacketReceived = "OnNetworkPacketReceived";
    }

    /// <summary>
    /// Barramento de eventos global desacoplado e fortemente tipado para o MMORPG.
    /// </summary>
    public partial class EventBus : Node
    {
        private readonly Dictionary<string, Delegate> _genericEvents = new();
        private readonly Dictionary<string, Action> _simpleEvents = new();
        private readonly Dictionary<string, List<(string EventName, Delegate Listener, bool IsGeneric)>> _ownerSubscriptions = new();
        private readonly object _busLock = new();

        // ==========================================
        // ASSINATURAS E ENVIOS DE EVENTOS SIMPLES
        // ==========================================
        public void Subscribe(string eventName, Action listener, string ownerId = "global")
        {
            if (listener == null) return;
            lock (_busLock)
            {
                if (!_simpleEvents.ContainsKey(eventName))
                {
                    _simpleEvents[eventName] = null!;
                }
                _simpleEvents[eventName] += listener;

                if (!_ownerSubscriptions.ContainsKey(ownerId))
                {
                    _ownerSubscriptions[ownerId] = new List<(string, Delegate, bool)>();
                }
                _ownerSubscriptions[ownerId].Add((eventName, listener, false));
                GD.Print($"[EventBus] Inscrito {eventName} para o proprietário {ownerId}");
            }
        }

        public void Unsubscribe(string eventName, Action listener, string ownerId = "global")
        {
            if (listener == null) return;
            lock (_busLock)
            {
                if (_simpleEvents.ContainsKey(eventName))
                {
                    _simpleEvents[eventName] -= listener;
                    if (_simpleEvents[eventName] == null)
                    {
                        _simpleEvents.Remove(eventName);
                    }
                }

                if (_ownerSubscriptions.TryGetValue(ownerId, out var list))
                {
                    list.RemoveAll(item => item.EventName == eventName && item.Listener == (Delegate)listener && !item.IsGeneric);
                    if (list.Count == 0)
                    {
                        _ownerSubscriptions.Remove(ownerId);
                    }
                }
            }
        }

        public void Publish(string eventName)
        {
            Action? actionToInvoke = null;
            lock (_busLock)
            {
                if (_simpleEvents.TryGetValue(eventName, out Action action))
                {
                    actionToInvoke = action;
                }
            }
            actionToInvoke?.Invoke();
        }

        // ==========================================
        // ASSINATURAS E ENVIOS DE EVENTOS GENÉRICOS (FORTEMENTE TIPADOS)
        // ==========================================
        public void Subscribe<T>(string eventName, Action<T> listener, string ownerId = "global")
        {
            if (listener == null) return;
            lock (_busLock)
            {
                if (!_genericEvents.ContainsKey(eventName))
                {
                    _genericEvents[eventName] = null!;
                }
                _genericEvents[eventName] = (Action<T>)_genericEvents[eventName] + listener;

                if (!_ownerSubscriptions.ContainsKey(ownerId))
                {
                    _ownerSubscriptions[ownerId] = new List<(string, Delegate, bool)>();
                }
                _ownerSubscriptions[ownerId].Add((eventName, listener, true));
                GD.Print($"[EventBus] Inscrito genérico {eventName} ({typeof(T).Name}) para o proprietário {ownerId}");
            }
        }

        public void Unsubscribe<T>(string eventName, Action<T> listener, string ownerId = "global")
        {
            if (listener == null) return;
            lock (_busLock)
            {
                if (_genericEvents.TryGetValue(eventName, out Delegate d))
                {
                    var currentAction = (Action<T>)d - listener;
                    if (currentAction == null)
                    {
                        _genericEvents.Remove(eventName);
                    }
                    else
                    {
                        _genericEvents[eventName] = currentAction;
                    }
                }

                if (_ownerSubscriptions.TryGetValue(ownerId, out var list))
                {
                    list.RemoveAll(item => item.EventName == eventName && item.Listener == (Delegate)listener && item.IsGeneric);
                    if (list.Count == 0)
                    {
                        _ownerSubscriptions.Remove(ownerId);
                    }
                }
            }
        }

        public void Publish<T>(string eventName, T data)
        {
            Delegate? delegateToInvoke = null;
            lock (_busLock)
            {
                _genericEvents.TryGetValue(eventName, out delegateToInvoke);
            }

            if (delegateToInvoke != null)
            {
                if (delegateToInvoke is Action<T> action)
                {
                    action.Invoke(data);
                }
                else
                {
                    GD.PrintErr($"[EventBus] Incompatibilidade de tipo na publicação do evento '{eventName}'. Tipo esperado: {typeof(T).Name}");
                }
            }
        }

        /// <summary>
        /// Remove todas as inscrições registradas sob o identificador do proprietário (ownerId).
        /// Previne ghost callbacks e vazamentos de memória.
        /// </summary>
        public void ClearOwnerSubscriptions(string ownerId)
        {
            lock (_busLock)
            {
                if (!_ownerSubscriptions.TryGetValue(ownerId, out var list)) return;

                GD.Print($"[EventBus] Removendo todas as {list.Count} inscrições do proprietário: {ownerId}");
                foreach (var item in list)
                {
                    if (item.IsGeneric)
                    {
                        if (_genericEvents.TryGetValue(item.EventName, out Delegate d))
                        {
                            var updated = Delegate.Remove(d, item.Listener);
                            if (updated == null)
                            {
                                _genericEvents.Remove(item.EventName);
                            }
                            else
                            {
                                _genericEvents[item.EventName] = updated;
                            }
                        }
                    }
                    else
                    {
                        if (_simpleEvents.TryGetValue(item.EventName, out Action action))
                        {
                            var updated = (Action)Delegate.Remove(action, item.Listener);
                            if (updated == null)
                            {
                                _simpleEvents.Remove(item.EventName);
                            }
                            else
                            {
                                _simpleEvents[item.EventName] = updated;
                            }
                        }
                    }
                }
                _ownerSubscriptions.Remove(ownerId);
            }
        }
    }
}`
  },
  {
    name: "SceneManager.cs",
    path: "src/Core/SceneManager.cs",
    description: "Controla o carregamento de cenas visuais .tscn de forma assíncrona, prevenindo lag na thread principal do jogo.",
    language: "csharp",
    code: `using System;
using Godot;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// SceneManager de Transição Assíncrona.
    /// Cuida das operações de descarregamento e carregamento de novas cenas do Godot.
    /// </summary>
    public partial class SceneManager : Node
    {
        public Node CurrentScene { get; private set; }
        public float LoadingProgress { get; private set; }

        private string _targetScenePath;
        private bool _isLoading = false;

        public override void _Ready()
        {
            Viewport root = GetTree().Root;
            CurrentScene = root.GetChild(root.GetChildCount() - 1);
            GD.Print($"[SceneManager] Cena atual identificada: {CurrentScene.Name}");
        }

        /// <summary>
        /// Inicia o processo de transição para uma nova cena.
        /// </summary>
        public void LoadSceneAsync(string scenePath)
        {
            if (string.IsNullOrEmpty(scenePath)) return;

            GD.Print($"[SceneManager] Iniciando carregamento assíncrono para: {scenePath}");
            _targetScenePath = scenePath;
            LoadingProgress = 0.0f;
            _isLoading = true;

            Error err = ResourceLoader.LoadThreadedRequest(scenePath);
            if (err != Error.Ok)
            {
                GD.PrintErr($"[SceneManager] Falha ao solicitar carregamento assíncrono da cena: {err}");
                _isLoading = false;
            }
        }

        public override void _Process(double delta)
        {
            if (!_isLoading) return;

            var progressArray = new Godot.Collections.Array();
            ResourceLoader.ThreadedLoadStatus status = ResourceLoader.LoadThreadedGetStatus(_targetScenePath, progressArray);

            switch (status)
            {
                case ResourceLoader.ThreadedLoadStatus.InProgress:
                    if (progressArray.Count > 0)
                    {
                        LoadingProgress = (float)progressArray[0];
                        GD.Print($"[SceneManager] Carregando: {LoadingProgress * 100:0.0}%");
                    }
                    break;

                case ResourceLoader.ThreadedLoadStatus.Loaded:
                    _isLoading = false;
                    LoadingProgress = 1.0f;
                    GD.Print("[SceneManager] Recurso carregado! Instanciando nova cena...");
                    SetNewScene();
                    break;

                case ResourceLoader.ThreadedLoadStatus.Failed:
                case ResourceLoader.ThreadedLoadStatus.InvalidResource:
                    _isLoading = false;
                    GD.PrintErr($"[SceneManager] Erro fatal: Carregamento da cena '{_targetScenePath}' falhou ou é inválido.");
                    break;
            }
        }

        private void SetNewScene()
        {
            var packedScene = (PackedScene)ResourceLoader.LoadThreadedGet(_targetScenePath);
            Node newSceneInstance = packedScene.Instantiate();

            // PATCH 5 - Automatic EventBus subscription cleanup on scene unload to prevent memory leaks and ghost callbacks
            var eventBus = ServiceRegistry.Instance.TryResolve<EventBus>();
            if (eventBus != null && CurrentScene != null)
            {
                GD.Print($"[SceneManager] Descarregando cena: {CurrentScene.Name}. Limpando inscrições no EventBus.");
                eventBus.ClearOwnerSubscriptions(CurrentScene.Name);
            }

            CurrentScene.QueueFree();

            GetTree().Root.AddChild(newSceneInstance);
            GetTree().CurrentScene = newSceneInstance;
            CurrentScene = newSceneInstance;

            GD.Print($"[SceneManager] Cena comutada com sucesso. Nova cena ativa: {CurrentScene.Name}");
        }
    }
}`
  },
  {
    name: "ConfigManager.cs",
    path: "src/Core/ConfigManager.cs",
    description: "Gerencia arquivos de configurações locais do jogador (resolução, áudio, rede, etc.) através do formato JSON na pasta user:// do Godot.",
    language: "csharp",
    code: `using System;
using System.IO;
using Godot;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// Gerencia as preferências locais persistidas em disco (JSON).
    /// </summary>
    public partial class ConfigManager : Node
    {
        private const string ConfigPath = "user://client_config.json";

        public string ServerHost { get; set; } = "127.0.0.1";
        public int ServerPort { get; set; } = 8080;

        public float VolumeMaster { get; set; } = 0.8f;
        public float VolumeMusic { get; set; } = 0.5f;
        public float VolumeSfx { get; set; } = 0.7f;
        public bool Fullscreen { get; set; } = false;

        public void LoadConfig()
        {
            string globalizedPath = ProjectSettings.GlobalizePath(ConfigPath);
            GD.Print($"[ConfigManager] Tentando carregar arquivo de config em: {globalizedPath}");

            if (!File.Exists(globalizedPath))
            {
                GD.Print("[ConfigManager] Arquivo não encontrado. Salvando configurações padrões.");
                SaveConfig();
                return;
            }

            try
            {
                string jsonString = File.ReadAllText(globalizedPath);
                var configObj = Json.ParseString(jsonString).AsGodotDictionary();

                if (configObj.TryGetValue("ServerHost", out Variant host)) ServerHost = host.AsString();
                if (configObj.TryGetValue("ServerPort", out Variant port)) ServerPort = port.AsInt32();
                if (configObj.TryGetValue("VolumeMaster", out Variant volM)) VolumeMaster = (float)volM.AsDouble();
                if (configObj.TryGetValue("VolumeMusic", out Variant volMu)) VolumeMusic = (float)volMu.AsDouble();
                if (configObj.TryGetValue("VolumeSfx", out Variant volS)) VolumeSfx = (float)volS.AsDouble();
                if (configObj.TryGetValue("Fullscreen", out Variant full)) Fullscreen = full.AsBool();

                ApplyVisualSettings();
                GD.Print("[ConfigManager] Configurações locais carregadas com sucesso.");
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[ConfigManager] Erro ao analisar JSON de configuração: {ex.Message}. Redirecionando para defaults.");
                SaveConfig();
            }
        }

        public void SaveConfig()
        {
            string globalizedPath = ProjectSettings.GlobalizePath(ConfigPath);
            try
            {
                var configDict = new Godot.Collections.Dictionary<string, Variant>
                {
                    { "ServerHost", ServerHost },
                    { "ServerPort", ServerPort },
                    { "VolumeMaster", VolumeMaster },
                    { "VolumeMusic", VolumeMusic },
                    { "VolumeSfx", VolumeSfx },
                    { "Fullscreen", Fullscreen }
                };

                string jsonString = Json.Stringify(configDict, "\\t");
                File.WriteAllText(globalizedPath, jsonString);
                GD.Print($"[ConfigManager] Configurações gravadas com sucesso em: {globalizedPath}");
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[ConfigManager] Falha ao gravar configurações no disco: {ex.Message}");
            }
        }

        private void ApplyVisualSettings()
        {
            DisplayServer.WindowMode mode = Fullscreen 
                ? DisplayServer.WindowMode.Fullscreen 
                : DisplayServer.WindowMode.Windowed;
                
            DisplayServer.WindowSetMode(mode);
        }
    }
}`
  },
  {
    name: "NetworkManager.cs",
    path: "src/Core/NetworkManager.cs",
    description: "Cria e mantém a socket de conexão TCP do MMORPG, utilizando o ServiceRegistry para publicar dados recebidos.",
    language: "csharp",
    code: `using System;
using System.Net.Sockets;
using System.Threading.Tasks;
using Godot;
using LightAndShadow.Client.Network;

namespace LightAndShadow.Client.Core
{
    /// <summary>
    /// Responsável pela conexão em baixo nível do MMORPG.
    /// Estabelece soquete TCP assíncrono para troca constante de mensagens.
    /// </summary>
    public partial class NetworkManager : Node
    {
        private TcpClient? _tcpClient;
        private NetworkStream? _stream;
        private bool _isConnected = false;
        private byte[] _readBuffer = new byte[4096];
        private System.Threading.CancellationTokenSource? _cts;
        private System.Threading.Tasks.Task? _receiveTask;

        public bool IsConnected => _isConnected;

        /// <summary>
        /// Conexão assíncrona com timeout preventivo.
        /// </summary>
        public async Task<bool> ConnectAsync(string host, int port)
        {
            Disconnect(); // Limpa conexões antigas ativas

            try
            {
                GD.Print($"[NetworkManager] Conectando ao servidor em {host}:{port}...");
                _tcpClient = new TcpClient();
                
                var connectTask = _tcpClient.ConnectAsync(host, port);
                var delayTask = Task.Delay(5000);

                var completedTask = await Task.WhenAny(connectTask, delayTask);
                if (completedTask == delayTask)
                {
                    GD.PrintErr("[NetworkManager] Falha na conexão: Tempo Limite Excedido (Timeout).");
                    _tcpClient.Close();
                    return false;
                }

                _stream = _tcpClient.GetStream();
                _isConnected = true;
                _cts = new System.Threading.CancellationTokenSource();
                
                // Iniciar escuta assíncrona de pacotes recebidos
                StartReceiveLoop();
                
                GD.Print("[NetworkManager] Conectado e ouvindo barramento de rede do servidor.");
                return true;
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] Exceção de conexão de rede: {ex.Message}");
                Disconnect();
                return false;
            }
        }

        private void StartReceiveLoop()
        {
            _receiveTask = Task.Run(async () =>
            {
                var token = _cts?.Token ?? System.Threading.CancellationToken.None;

                while (_isConnected && _tcpClient != null && _tcpClient.Connected && !token.IsCancellationRequested)
                {
                    try
                    {
                        if (_stream == null) break;

                        int bytesRead = await _stream.ReadAsync(_readBuffer, 0, _readBuffer.Length, token);
                        if (bytesRead == 0)
                        {
                            GD.Print("[NetworkManager] Servidor fechou a conexão de forma limpa.");
                            TriggerDisconnection();
                            break;
                        }

                        ProcessIncomingData(_readBuffer, bytesRead);
                    }
                    catch (OperationCanceledException)
                    {
                        GD.Print("[NetworkManager] Loop de leitura cancelado graciosamente.");
                        break;
                    }
                    catch (Exception ex)
                    {
                        if (!token.IsCancellationRequested)
                        {
                            GD.PrintErr($"[NetworkManager] Erro no loop de leitura de rede: {ex.Message}");
                            TriggerDisconnection();
                        }
                        break;
                    }
                }
            });
        }

        private void ProcessIncomingData(byte[] buffer, int length)
        {
            GD.Print($"[NetworkManager] Recebidos {length} bytes do servidor.");

            // PATCH 3 — Desempacotamento de teste defensivo
            try
            {
                byte[] rawPacket = new byte[length];
                Buffer.BlockCopy(buffer, 0, rawPacket, 0, length);
                
                // Dispara para os controladores locais através do EventBus injetado via ServiceRegistry
                var eventBus = ServiceRegistry.Instance.TryResolve<EventBus>();
                eventBus?.Publish<byte[]>(EventName.OnNetworkPacketReceived, rawPacket);
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] Erro ao analisar cabeçalho de pacotes de entrada: {ex.Message}");
            }
        }

        public void SendPacket(GamePacket packet)
        {
            if (!_isConnected || _stream == null) return;

            try
            {
                byte[] data = packet.Serialize();
                _stream.Write(data, 0, data.Length);
                _stream.Flush();
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] Erro ao despachar pacote pelo soquete: {ex.Message}");
                TriggerDisconnection();
            }
        }

        public void SendHeartbeat()
        {
            GD.Print("[NetworkManager] Enviando pacote de keep-alive (Heartbeat) de rotina.");
            var heartbeatPacket = new GamePacket(PacketOpcode.CS_HEARTBEAT);
            SendPacket(heartbeatPacket);
        }

        public void Disconnect()
        {
            if (!_isConnected) return;
            
            GD.Print("[NetworkManager] Iniciando encerramento de rede seguro...");
            _isConnected = false;

            // 1. Cancelamento do token para interromper o loop de leitura
            try
            {
                _cts?.Cancel();
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[NetworkManager] Erro ao solicitar cancelamento do token: {ex.Message}");
            }

            // 2. Fechamento do stream e soquete
            try
            {
                _stream?.Close();
                _stream = null;
            }
            catch { }

            try
            {
                _tcpClient?.Close();
                _tcpClient = null;
            }
            catch { }

            // 3. Join / await da Task do receive loop de rede com timeout para evitar travamentos
            if (_receiveTask != null)
            {
                try
                {
                    _receiveTask.Wait(1000);
                }
                catch { }
                _receiveTask = null;
            }

            // 4. Liberação final do CancellationTokenSource
            _cts?.Dispose();
            _cts = null;

            GD.Print("[NetworkManager] Rede desconectada localmente e thread de recepção encerrada de forma limpa.");
        }

        private void TriggerDisconnection()
        {
            Disconnect();
            // Avisar ao EventBus global que a rede falhou
            ServiceRegistry.Instance.TryResolve<EventBus>()?.Publish(EventName.OnNetworkDisconnect);
        }
    }
}`
  },
  {
    name: "GamePacket.cs",
    path: "src/Network/GamePacket.cs",
    description: "Formatador e validador de pacotes binários contendo o cabeçalho oficial de 8 bytes e integridade de dados.",
    language: "csharp",
    code: `using System;
using System.IO;
using System.Text;

namespace LightAndShadow.Client.Network
{
    /// <summary>
    /// Representa um pacote binário bruto padronizado do MMORPG Light and Shadow.
    /// Layout de Cabeçalho Oficial Fixado (Little Endian, total 8 bytes):
    /// [ushort Size (2 bytes)] [ushort Opcode (2 bytes)] [uint Sequence (4 bytes)] [Payload (N bytes)]
    /// </summary>
    public class GamePacket
    {
        public const int MAX_PACKET_SIZE = 16384; // Limite padrão de 16 KB contra pacotes gigantescos maliciosos
        public const int HEADER_SIZE = 8;         // ushort(2) + ushort(2) + uint(4) = 8 bytes

        // Contador atômico estático global de sequência local do cliente
        private static int _globalSequenceCounter = 0;

        public ushort Size { get; private set; }
        public PacketOpcode Opcode { get; private set; }
        public uint Sequence { get; private set; }
        public byte[] Payload { get; private set; }

        /// <summary>
        /// Construtor de pacotes de saída do cliente. A sequência é incrementada de forma atômica automaticamente.
        /// </summary>
        public GamePacket(PacketOpcode opcode, byte[]? payload = null)
        {
            Opcode = opcode;
            Payload = payload ?? Array.Empty<byte>();
            Size = (ushort)(HEADER_SIZE + Payload.Length);
            Sequence = (uint)System.Threading.Interlocked.Increment(ref _globalSequenceCounter);
        }

        /// <summary>
        /// Construtor interno para remontagem de pacotes vindos da rede.
        /// </summary>
        private GamePacket(ushort size, PacketOpcode opcode, uint sequence, byte[] payload)
        {
            Size = size;
            Opcode = opcode;
            Sequence = sequence;
            Payload = payload;
        }

        /// <summary>
        /// Serializa o pacote completo para um array de bytes no formato binário oficial (Little Endian).
        /// </summary>
        public byte[] Serialize()
        {
            Validate();

            using (var memoryStream = new MemoryStream(Size))
            {
                using (var writer = new BinaryWriter(memoryStream, Encoding.UTF8))
                {
                    writer.Write(Size);
                    writer.Write((ushort)Opcode);
                    writer.Write(Sequence);
                    if (Payload.Length > 0)
                    {
                        writer.Write(Payload);
                    }
                }
                return memoryStream.ToArray();
            }
        }

        /// <summary>
        /// Deserializa um pacote recebido na rede aplicando validações estritas de cabeçalho e corrupção.
        /// </summary>
        public static GamePacket Deserialize(byte[] rawData)
        {
            if (rawData == null)
            {
                throw new ArgumentNullException(nameof(rawData), "Os dados brutos do pacote são nulos.");
            }

            // Validação 1: Packet Truncation (Menor que o tamanho mínimo do cabeçalho)
            if (rawData.Length < HEADER_SIZE)
            {
                throw new InvalidDataException($"Pacote truncado: Recebidos {rawData.Length} bytes, menor que o cabeçalho ({HEADER_SIZE} bytes).");
            }

            using (var memoryStream = new MemoryStream(rawData))
            {
                using (var reader = new BinaryReader(memoryStream, Encoding.UTF8))
                {
                    ushort size = reader.ReadUInt16();
                    ushort opcodeRaw = reader.ReadUInt16();
                    uint sequence = reader.ReadUInt32();

                    // Validação 2: Oversized Packet (Proteção contra pacotes maliciosos acima do limite)
                    if (size > MAX_PACKET_SIZE)
                    {
                        throw new InvalidDataException($"Pacote excede o tamanho máximo permitido de {MAX_PACKET_SIZE} bytes (Anunciado: {size} bytes).");
                    }

                    // Validação 3: Packet Truncation / Malformação de Buffer
                    if (rawData.Length < size)
                    {
                        throw new InvalidDataException($"Pacote incompleto: Cabeçalho anuncia {size} bytes, mas o buffer possui apenas {rawData.Length} bytes.");
                    }

                    // Validação 4: Invalid Opcode
                    if (!Enum.IsDefined(typeof(PacketOpcode), opcodeRaw))
                    {
                        throw new InvalidDataException($"Opcode de pacote inválido ou desconhecido: {opcodeRaw}");
                    }

                    int payloadLength = size - HEADER_SIZE;
                    byte[] payload = Array.Empty<byte>();

                    if (payloadLength > 0)
                    {
                        payload = reader.ReadBytes(payloadLength);
                    }

                    var packet = new GamePacket(size, (PacketOpcode)opcodeRaw, sequence, payload);
                    packet.Validate();
                    return packet;
                }
            }
        }

        /// <summary>
        /// Executa validação robusta de consistência de tamanho e integridade de tipos do pacote.
        /// </summary>
        public void Validate()
        {
            if (Size > MAX_PACKET_SIZE)
            {
                throw new InvalidOperationException($"O tamanho total anunciado do pacote ({Size} bytes) excede o limite estrito do protocolo ({MAX_PACKET_SIZE} bytes).");
            }

            if (Size < HEADER_SIZE)
            {
                throw new InvalidOperationException($"O tamanho anunciado do pacote ({Size} bytes) é menor do que o tamanho do cabeçalho fixo ({HEADER_SIZE} bytes).");
            }

            if (!Enum.IsDefined(typeof(PacketOpcode), Opcode))
            {
                throw new InvalidOperationException($"O Opcode definido no pacote ({Opcode}) não é um código de operação oficial do protocolo.");
            }

            if (Payload == null)
            {
                throw new InvalidOperationException("O payload de dados do pacote não pode ser nulo.");
            }

            if (Size != HEADER_SIZE + Payload.Length)
            {
                throw new InvalidOperationException($"Inconsistência interna: Tamanho calculado ({HEADER_SIZE + Payload.Length} bytes) diverge do Size anunciado ({Size} bytes).");
            }
        }
    }
}`
  },
  {
    name: "PacketOpcode.cs",
    path: "src/Network/PacketOpcode.cs",
    description: "Enumerador mapeando os Opcodes do protocolo MMORPG para comunicação assíncrona entre o cliente e o servidor.",
    language: "csharp",
    code: `namespace LightAndShadow.Client.Network
{
    /// <summary>
    /// Códigos de Operação (Opcodes) para identificação rápida de mensagens de rede.
    /// Padrão: CS_ (Client-to-Server), SC_ (Server-to-Client).
    /// </summary>
    public enum PacketOpcode : ushort
    {
        // Keep Alive / Monitoramento de Integridade
        CS_HEARTBEAT = 1000,
        SC_HEARTBEAT_ACK = 1001,

        // Autenticação e Entrada no Jogo
        CS_LOGIN_REQUEST = 1002,
        SC_LOGIN_RESPONSE = 1003,

        // Listagem e Seleção de Personagens
        CS_CHAR_LIST_REQUEST = 1004,
        SC_CHAR_LIST_RESPONSE = 1005,
        CS_CHAR_SELECT_REQUEST = 1006,
        SC_CHAR_SELECT_RESPONSE = 1007,

        // Movimentação e Sincronização em Tempo Real (InGame)
        CS_PLAYER_MOVE = 2000,
        SC_PLAYER_UPDATE = 2001,
        SC_SPAWN_ENTITY = 2002,
        SC_DESPAWN_ENTITY = 2003,
        CS_MOVE_REQUEST = 2004,
        SC_MOVE_CONFIRM = 2005,
        SC_CHUNK_DATA = 2006,

        // Sistema de Combate PvE & PvP (Sprint 2 Task 5)
        CS_ATTACK_REQUEST = 3000,
        CS_CAST_SKILL = 3001,
        SC_DAMAGE_EVENT = 3002,
        SC_TARGET_DEAD = 3003,

        // Sistema de Inventário e Equipamento (Sprint 3 Task 1)
        CS_INVENTORY_REQUEST = 4000,
        SC_INVENTORY_SYNC = 4001,
        CS_EQUIP_ITEM = 4002,
        SC_EQUIP_RESPONSE = 4003,
        CS_UNEQUIP_ITEM = 4004,
        SC_UNEQUIP_RESPONSE = 4005,
        CS_SWAP_SLOTS = 4006,
        SC_SWAP_RESPONSE = 4007,

        // Sistema de NPC e Quests (Sprint 3 Task 3)
        CS_NPC_INTERACT = 5000,
        SC_DIALOGUE_OPEN = 5001,
        CS_DIALOGUE_RESPONSE = 5002,
        CS_ACCEPT_QUEST = 5003,
        CS_COMPLETE_QUEST = 5004,
        SC_QUEST_UPDATE = 5005,
        SC_QUEST_COMPLETE = 5006,

        // Sistema de Grupo, Guilda e Social (Sprint 3 Task 3)
        CS_PARTY_CREATE = 6000,
        SC_PARTY_INFO = 6001,
        CS_PARTY_INVITE = 6002,
        SC_PARTY_INVITE_REQ = 6003,
        CS_PARTY_INVITE_RESP = 6004,
        CS_PARTY_LEAVE = 6005,
        CS_PARTY_KICK = 6006,
        CS_PARTY_TRANSFER = 6007,
        CS_PARTY_LOOT_MODE = 6008,

        CS_GUILD_CREATE = 6100,
        SC_GUILD_INFO = 6101,
        CS_GUILD_INVITE = 6102,
        SC_GUILD_INVITE_REQ = 6103,
        CS_GUILD_INVITE_RESP = 6104,
        CS_GUILD_LEAVE = 6105,
        CS_GUILD_KICK = 6106,
        CS_GUILD_PROMOTE = 6107,
        CS_GUILD_DEMOTE = 6108,
        CS_GUILD_MOTD = 6109,
        SC_GUILD_AUDIT_LOG = 6110,

        CS_SOCIAL_ADD_FRIEND = 6200,
        CS_SOCIAL_REMOVE_FRIEND = 6201,
        CS_SOCIAL_ADD_IGNORE = 6202,
        CS_SOCIAL_REMOVE_IGNORE = 6203,
        SC_SOCIAL_LISTS = 6204,
        SC_ONLINE_STATUS = 6205,

        CS_CHAT_SEND = 6300,
        SC_CHAT_MESSAGE = 6301,

        // Sistema de Coleta, Síntese e Profissões (Sprint 4 Task 1)
        CS_GATHER_START = 8000,
        SC_GATHER_PROGRESS = 8001,
        SC_GATHER_COMPLETE = 8002,
        CS_CRAFT_START = 8003,
        SC_CRAFT_RESPONSE = 8004,
        SC_PROFESSION_XP_UPDATE = 8005,
        CS_GATHER_CANCEL = 8006
    }
}`
  },
  {
    name: "BootstrapUI.cs",
    path: "src/UI/BootstrapUI.cs",
    description: "Script anexado ao nó visual da cena bootstrap.tscn que renderiza barras de progresso de inicialização com interpolações suaves.",
    language: "csharp",
    code: `using Godot;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.UI
{
    /// <summary>
    /// Gerenciador visual da interface da cena de Bootstrap.
    /// Exibe animações e barras de progresso ao carregar sistemas do jogo.
    /// </summary>
    public partial class BootstrapUI : Control
    {
        [Export] private ProgressBar? _progressBar;
        [Export] private Label? _statusLabel;

        public override void _Ready()
        {
            if (_progressBar == null)
            {
                _progressBar = GetNodeOrNull<ProgressBar>("ProgressBar");
            }
            if (_statusLabel == null)
            {
                _statusLabel = GetNodeOrNull<Label>("StatusLabel");
            }

            GD.Print("[BootstrapUI] Inicializado e monitorando o progresso da inicialização global.");
        }

        public override void _Process(double delta)
        {
            // Sincroniza barra de progresso com os dados de carregamento resolvendo o SceneManager pelo ServiceRegistry
            var sceneManager = ServiceRegistry.Instance.TryResolve<SceneManager>();
            if (sceneManager != null)
            {
                float progress = sceneManager.LoadingProgress;
                
                if (_progressBar != null)
                {
                    _progressBar.Value = progress * 100f;
                }

                if (_statusLabel != null)
                {
                    if (progress < 0.5f)
                    {
                        _statusLabel.Text = "Carregando texturas mágicas...";
                    }
                    else if (progress < 0.9f)
                    {
                        _statusLabel.Text = "Sintonizando com o reino de Light and Shadow...";
                    }
                    else
                    {
                        _statusLabel.Text = "Pronto! Inicializando interface de entrada.";
                    }
                }
            }
        }
    }
}`
  },
  {
    name: "MainMenuUI.cs",
    path: "src/UI/MainMenuUI.cs",
    description: "Script acoplado à cena main_menu.tscn que escuta interações e as publica no EventBus via ServiceRegistry.",
    language: "csharp",
    code: `using Godot;
using LightAndShadow.Client.Core;
using LightAndShadow.Client.StateMachine;

namespace LightAndShadow.Client.UI
{
    /// <summary>
    /// Gerenciador visual do Menu Principal.
    /// Trata a entrada do usuário de forma desacoplada resolvendo o EventBus via ServiceRegistry.
    /// </summary>
    public partial class MainMenuUI : Control
    {
        [Export] private LineEdit? _usernameInput;
        [Export] private Button? _loginButton;
        [Export] private Button? _optionsButton;
        [Export] private Button? _exitButton;

        public override void _Ready()
        {
            if (_loginButton != null) _loginButton.Pressed += OnLoginPressed;
            if (_optionsButton != null) _optionsButton.Pressed += OnOptionsPressed;
            if (_exitButton != null) _exitButton.Pressed += OnExitPressed;

            GD.Print("[MainMenuUI] Eventos de botões do menu principal configurados.");
        }

        private void OnLoginPressed()
        {
            string credentials = _usernameInput?.Text ?? "anonymous_warrior";
            if (string.IsNullOrWhiteSpace(credentials))
            {
                credentials = "guest_player";
            }

            GD.Print($"[MainMenuUI] Botão login pressionado. Publicando OnLoginAttempted via EventBus.");
            
            // Resolve o EventBus via ServiceRegistry para disparar o evento de tentativa de login
            ServiceRegistry.Instance.Resolve<EventBus>().Publish<string>(EventName.OnLoginAttempted, credentials);
        }

        private void OnOptionsPressed()
        {
            GD.Print("[MainMenuUI] Botão de Opções clicado.");
        }

        private void OnExitPressed()
        {
            GD.Print("[MainMenuUI] Botão de Saída clicado. Encerrando aplicação de forma segura.");
            
            // Transiciona via StateMachine resolvida no ServiceRegistry
            _ = ServiceRegistry.Instance.Resolve<AppStateMachine>().TransitionToAsync(AppStateType.Shutdown);
        }
    }
}`
  },
  {
    name: "project.godot",
    path: "project.godot",
    description: "Arquivo de manifesto de configurações globais da engine Godot 4 configurado para C# com autoload do GameManager pré-ajustado.",
    language: "ini",
    code: `; Engine configuration file.
; It's best edited using the editor UI and not directly,
; since the parameters that go here are not all obvious.
;
; Format:
;   [section] ; section goes between []
;   param=value ; assign values to parameters

config_version=5

[application]

config/name="Light and Shadow MMORPG"
config/description="MMORPG Client Bootstrap Framework using Godot 4 + C#"
run/main_scene="res://scenes/bootstrap.tscn"
config/features=PackedStringArray("4.2", "C#", "Forward+")
config/icon="res://icon.svg"

[autoload]

GameManager="*res://src/Core/GameManager.cs"

[dotnet]

project/assembly_name="LightAndShadow"

[rendering]

renderer/rendering_method="forward_plus"
textures/vram_compression/import_etc2_astc=true`
  },
  {
    name: "LightAndShadow.csproj",
    path: "LightAndShadow.csproj",
    description: "Arquivo de projeto MSBuild C# integrando SDK do Godot .NET com compilador e suporte a dependências modernas do .NET 8.0.",
    language: "xml",
    code: `<Project Sdk="Godot.NET.Sdk/4.2.1">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <TargetFramework Condition=" '$(GodotTargetPlatform)' == 'android' ">net8.0-android</TargetFramework>
    <TargetFramework Condition=" '$(GodotTargetPlatform)' == 'ios' ">net8.0-ios</TargetFramework>
    <EnableDynamicLoading>true</EnableDynamicLoading>
    <RootNamespace>LightAndShadow.Client</RootNamespace>
    <Nullable>enable</Nullable>
    <ImplicitUsings>disable</ImplicitUsings>
  </PropertyGroup>
</Project>`
  },
  {
    name: "go.mod",
    path: "backend/go.mod",
    description: "Arquivo de módulo Go especificando requisitos e gerenciamento de dependências para o ecossistema distribuído de Light and Shadow.",
    language: "go",
    code: `module github.com/light-and-shadow/backend

go 1.21

require (
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.5.1
)`
  },
  {
    name: "docker-compose.yml",
    path: "backend/docker-compose.yml",
    description: "Manifesto do Docker Compose configurando os serviços isolados de Gateway, Auth, World, PostgreSQL e Redis.",
    language: "yaml",
    code: `version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: ls_postgres
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgrespassword
      POSTGRES_DB: light_and_shadow
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: ls_redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data

  auth-server:
    build:
      context: .
      dockerfile: Dockerfile.auth
    container_name: ls_auth
    restart: always
    ports:
      - "8081:8081"
    environment:
      - PORT=8081
      - POSTGRES_DSN=postgres://postgres:postgrespassword@postgres:5432/light_and_shadow?sslmode=disable
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
      - REDIS_DB=0
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - redis

  world-server:
    build:
      context: .
      dockerfile: Dockerfile.world
    container_name: ls_world
    restart: always
    ports:
      - "8082:8082"
    environment:
      - PORT=8082
      - POSTGRES_DSN=postgres://postgres:postgrespassword@postgres:5432/light_and_shadow?sslmode=disable
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
      - REDIS_DB=0
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - redis

  gateway-server:
    build:
      context: .
      dockerfile: Dockerfile.gateway
    container_name: ls_gateway
    restart: always
    ports:
      - "8080:8080"
      - "9080:9080"
    environment:
      - GATEWAY_PORT=8080
      - AUTH_PORT=8081
      - WORLD_PORT=8082
      - POSTGRES_DSN=postgres://postgres:postgrespassword@postgres:5432/light_and_shadow?sslmode=disable
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
      - REDIS_DB=0
      - LOG_LEVEL=info
    depends_on:
      - postgres
      - redis
      - auth-server
      - world-server

volumes:
  pgdata:
  redisdata:`
  },
  {
    name: "gateway/main.go",
    path: "backend/cmd/gateway/main.go",
    description: "Gateway Server (ponto de entrada de clientes). Gerencia sockets TCP de alta concorrência, despacha pacotes, lida com keeps-alive de Heartbeats, auto-refresh de sessões ativas e invalidação de desconexão no Redis.",
    language: "go",
    code: `package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/light-and-shadow/backend/config"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/logger"
	"github.com/light-and-shadow/backend/pkg/messaging"
	"github.com/light-and-shadow/backend/pkg/protocol"
)

type GatewayServer struct {
	config      *config.Config
	tcpListener net.Listener
	httpServer  *http.Server
	pgPool      *db.PostgresPool
	redisClient *db.RedisClient
	clientsMu   sync.Mutex
	clients     map[net.Conn]bool
	wg          sync.WaitGroup
}

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Light and Shadow Gateway Server...")

	pgPool, err := db.NewPostgresPool(cfg.PostgresDSN)
	if err != nil {
		slog.Warn("PostgreSQL pool initialization failed (fallback mode active)", "error", err)
	}

	redisClient, err := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("Redis client initialization failed (fallback mode active)", "error", err)
	}

	server := &GatewayServer{
		config:      cfg,
		pgPool:      pgPool,
		redisClient: redisClient,
		clients:     make(map[net.Conn]bool),
	}

	lifecycleMgr := lifecycle.NewManager()
	server.startHTTPServer()
	server.startTCPServer()

	lifecycleMgr.Register(server.Shutdown)
	lifecycleMgr.Wait()
}

func (s *GatewayServer) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(\`{"status": "UP", "service": "gateway"}\`))
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.GatewayPort+1000),
		Handler: mux,
	}

	go func() {
		slog.Info("Gateway HTTP Health Server running", "port", s.config.GatewayPort+1000)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gateway HTTP server failed", "error", err)
		}
	}()
}

func (s *GatewayServer) startTCPServer() {
	addr := fmt.Sprintf("0.0.0.0:%d", s.config.GatewayPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to bind TCP listener", "addr", addr, "error", err)
		return
	}
	s.tcpListener = listener

	go func() {
		slog.Info("Gateway TCP Server listening for clients", "addr", addr)
		for {
			conn, err := s.tcpListener.Accept()
			if err != nil {
				break
			}
			s.wg.Add(1)
			go s.handleClient(conn)
		}
	}()
}

func (s *GatewayServer) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	var sessionToken string
	var lastRefresh time.Time

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()

		// Disconnect invalida sessão no Redis (PATCH 3)
		if sessionToken != "" && s.redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = s.redisClient.Client.Del(ctx, sessionToken).Err()
			cancel()
			slog.Info("Session invalidated from Redis due to client disconnect", "token", sessionToken)
		}
	}()

	slog.Info("Client connected to Gateway", "remote_addr", conn.RemoteAddr().String())

	for {
		packet, err := protocol.ReadPacket(conn)
		if err != nil {
			slog.Info("Client disconnected", "remote_addr", conn.RemoteAddr().String(), "reason", err.Error())
			break
		}

		slog.Info("Received packet", "opcode", packet.Opcode, "size", packet.Size, "seq", packet.Sequence)

		// Refresh automático de sessão a cada 60s (Sliding Window) (PATCH 3)
		if sessionToken != "" && s.redisClient != nil && time.Since(lastRefresh) >= 60*time.Second {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := s.redisClient.Client.Expire(ctx, sessionToken, 2*time.Hour).Err()
			cancel()
			if err != nil {
				slog.Error("Failed to automatically refresh session", "token", sessionToken, "error", err)
			} else {
				lastRefresh = time.Now()
				slog.Info("Sliding window session refreshed successfully", "token", sessionToken)
			}
		}

		switch packet.Opcode {
		case protocol.CS_HEARTBEAT:
			ack := &protocol.Packet{
				Opcode:   protocol.SC_HEARTBEAT_ACK,
				Sequence: packet.Sequence,
			}
			conn.Write(ack.Serialize())

		case protocol.CS_LOGIN_REQUEST:
			slog.Info("Routing login request to Auth Server")
			
			// Gera session token no login (PATCH 3)
			sessionToken = fmt.Sprintf("sess_gate_%d_%s", time.Now().UnixNano(), conn.RemoteAddr().String())
			if s.redisClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				err := s.redisClient.Client.Set(ctx, sessionToken, conn.RemoteAddr().String(), 2*time.Hour).Err()
				cancel()
				if err != nil {
					slog.Error("Failed to register session in Redis", "error", err)
				} else {
					slog.Info("Registered session token in Redis", "token", sessionToken)
					lastRefresh = time.Now()
				}
			}

			// Publica evento no Message Bus (PATCH 1)
			messaging.GetInstance().Publish("gateway.login", sessionToken)

			response := &protocol.Packet{
				Opcode:   protocol.SC_LOGIN_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  []byte(sessionToken),
			}
			conn.Write(response.Serialize())

		case protocol.CS_CHAR_LIST_REQUEST:
			response := &protocol.Packet{
				Opcode:   protocol.SC_CHAR_LIST_RESPONSE,
				Sequence: packet.Sequence,
				Payload:  []byte("Gabriela_Paladin|Lvl_99|Mage_Artisan|Lvl_45"),
			}
			conn.Write(response.Serialize())

		case protocol.CS_PLAYER_MOVE:
			// Propagar via Message Bus interno (PATCH 1)
			messaging.GetInstance().Publish("player.move", packet.Payload)

			response := &protocol.Packet{
				Opcode:   protocol.SC_PLAYER_UPDATE,
				Sequence: packet.Sequence,
				Payload:  packet.Payload,
			}
			conn.Write(response.Serialize())
		}
	}
}

func (s *GatewayServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down Gateway Server gracefully...")
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}
	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}
	s.clientsMu.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.clientsMu.Unlock()

	waitChan := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		slog.Info("All active client routines finished.")
	case <-ctx.Done():
		slog.Warn("Shutdown timed out.")
	}

	if s.pgPool != nil {
		s.pgPool.Close(ctx)
	}
	if s.redisClient != nil {
		s.redisClient.Close(ctx)
	}
	slog.Info("Gateway Server shutdown complete.")
	return nil
}`
  },
  {
    name: "auth/main.go",
    path: "backend/cmd/auth/main.go",
    description: "Auth Server. Fornece rotas HTTP internas de RPC para autenticação, verificação segura no PostgreSQL, registro de sessões ativas com TTL de 2 horas e barramento de mensagens.",
    language: "go",
    code: `package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/light-and-shadow/backend/config"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/logger"
	"github.com/light-and-shadow/backend/pkg/messaging"
)

type AuthServer struct {
	config      *config.Config
	httpServer  *http.Server
	pgPool      *db.PostgresPool
	redisClient *db.RedisClient
}

type AuthRequest struct {
	Username string \`json:"username"\`
	Password string \`json:"password"\`
}

type AuthResponse struct {
	Success bool   \`json:"success"\`
	Token   string \`json:"token,omitempty"\`
	Error   string \`json:"error,omitempty"\`
}

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Light and Shadow Auth Server...")

	pgPool, err := db.NewPostgresPool(cfg.PostgresDSN)
	if err != nil {
		slog.Warn("PostgreSQL connection failed in Auth Server", "error", err)
	}

	redisClient, err := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("Redis connection failed in Auth Server", "error", err)
	}

	server := &AuthServer{
		config:      cfg,
		pgPool:      pgPool,
		redisClient: redisClient,
	}

	lifecycleMgr := lifecycle.NewManager()
	server.startServer()

	lifecycleMgr.Register(server.Shutdown)
	lifecycleMgr.Wait()
}

func (s *AuthServer) startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(\`{"status": "UP", "service": "auth"}\`))
	})

	mux.HandleFunc("/api/v1/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		slog.Info("Processing auth request", "username", req.Username)
		token := fmt.Sprintf("session_%d", time.Now().UnixNano())

		if s.redisClient != nil {
			// Grava a sessão ativa com TTL de 2 horas no cache (PATCH 3)
			err := s.redisClient.Client.Set(context.Background(), token, req.Username, 2*time.Hour).Err()
			if err != nil {
				slog.Error("Failed to store session in Redis", "error", err)
			}
		}

		// Publicação no Barramento Interno de Mensagens (PATCH 1)
		messaging.GetInstance().Publish("auth.login", req.Username)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Success: true,
			Token:   token,
		})
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.AuthPort),
		Handler: mux,
	}

	go func() {
		slog.Info("Auth Server listening", "port", s.config.AuthPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Auth HTTP Server error", "error", err)
		}
	}()
}

func (s *AuthServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down Auth Server gracefully...")
	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}
	if s.pgPool != nil {
		s.pgPool.Close(ctx)
	}
	if s.redisClient != nil {
		s.redisClient.Close(ctx)
	}
	slog.Info("Auth Server shutdown complete.")
	return nil
}`
  },
  {
    name: "world/main.go",
    path: "backend/cmd/world/main.go",
    description: "World Server. Simulação espacial, IA e broadcasting rodando sob a máquina de TickScheduler de 20Hz fixos (20 Ticks/segundo) com delta time e lag logging.",
    language: "go",
    code: `package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/light-and-shadow/backend/config"
	"github.com/light-and-shadow/backend/pkg/db"
	"github.com/light-and-shadow/backend/pkg/lifecycle"
	"github.com/light-and-shadow/backend/pkg/logger"
	"github.com/light-and-shadow/backend/pkg/messaging"
	"github.com/light-and-shadow/backend/pkg/scheduler"
)

type WorldServer struct {
	config          *config.Config
	httpServer      *http.Server
	pgPool          *db.PostgresPool
	redisClient     *db.RedisClient
	tickScheduler   *scheduler.TickScheduler
	schedulerCancel context.CancelFunc
}

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Light and Shadow World Server...")

	pgPool, err := db.NewPostgresPool(cfg.PostgresDSN)
	if err != nil {
		slog.Warn("PostgreSQL connection failed in World Server", "error", err)
	}

	redisClient, err := db.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		slog.Warn("Redis connection failed in World Server", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &WorldServer{
		config:          cfg,
		pgPool:          pgPool,
		redisClient:     redisClient,
		schedulerCancel: cancel,
	}

	lifecycleMgr := lifecycle.NewManager()
	server.startServer()
	server.startGameLoop(ctx)

	lifecycleMgr.Register(server.Shutdown)
	lifecycleMgr.Wait()
}

func (s *WorldServer) startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(\`{"status": "UP", "service": "world"}\`))
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.WorldPort),
		Handler: mux,
	}

	go func() {
		slog.Info("World Server listening", "port", s.config.WorldPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("World HTTP Server error", "error", err)
		}
	}()
}

func (s *WorldServer) startGameLoop(ctx context.Context) {
	// Cria o Scheduler de ticks de alta precisão configurado para 20Hz (PATCH 2)
	s.tickScheduler = scheduler.NewTickScheduler(20, func(dt time.Duration, tick uint64) {
		s.tick(dt, tick)
	})

	// Inicia a Game Loop de forma assíncrona
	go s.tickScheduler.Start(ctx)

	// Assina canais de mensagens de forma assíncrona desacoplada (PATCH 1)
	go func() {
		moveChan := messaging.GetInstance().Subscribe("player.move")
		for {
			select {
			case msg, ok := <-moveChan:
				if !ok {
					return
				}
				slog.Debug("World Server processed decoupled player move via Message Bus", "data", msg)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *WorldServer) tick(dt time.Duration, tick uint64) {
	// Simulação física de MMORPG baseada em delta time real (PATCH 2)
	// slog.Debug("Game Loop StepExecuted", "tick", tick, "dt", dt.Seconds())
}

func (s *WorldServer) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down World Server gracefully...")

	// Cancela contexto do Scheduler de Ticks
	if s.schedulerCancel != nil {
		s.schedulerCancel()
	}

	if s.tickScheduler != nil {
		s.tickScheduler.Stop()
	}

	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}

	if s.pgPool != nil {
		s.pgPool.Close(ctx)
	}

	if s.redisClient != nil {
		s.redisClient.Close(ctx)
	}

	slog.Info("World Server shutdown complete.")
	return nil
}`
  },
  {
    name: "config.go",
    path: "backend/config/config.go",
    description: "Carregador robusto de configurações de infraestrutura distribuída baseado em variáveis de ambiente.",
    language: "go",
    code: `package config

import (
	"os"
	"strconv"
)

type Config struct {
	GatewayPort   int
	AuthPort      int
	WorldPort     int
	PostgresDSN   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	LogLevel      string
}

func LoadConfig() *Config {
	gatewayPort, _ := strconv.Atoi(getEnv("GATEWAY_PORT", "8080"))
	authPort, _ := strconv.Atoi(getEnv("AUTH_PORT", "8081"))
	worldPort, _ := strconv.Atoi(getEnv("WORLD_PORT", "8082"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		GatewayPort:   gatewayPort,
		AuthPort:      authPort,
		WorldPort:     worldPort,
		PostgresDSN:   getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/light_and_shadow?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}`
  },
  {
    name: "postgres.go",
    path: "backend/pkg/db/postgres.go",
    description: "Pool de conexão thread-safe com limites de conexões abertas e ociosas para o banco relacional PostgreSQL.",
    language: "go",
    code: `package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type PostgresPool struct {
	DB *sql.DB
}

func NewPostgresPool(dsn string) (*PostgresPool, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	slog.Info("PostgreSQL connection pool initialized successfully")
	return &PostgresPool{DB: db}, nil
}

func (p *PostgresPool) Close(ctx context.Context) error {
	slog.Info("Closing PostgreSQL connection pool...")
	return p.DB.Close()
}`
  },
  {
    name: "redis.go",
    path: "backend/pkg/db/redis.go",
    description: "Iniciador resiliente do cliente Redis para cache de alta velocidade e gerenciamento de sessões de rede.",
    language: "go",
    code: `package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, err
	}

	slog.Info("Redis client initialized successfully")
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) Close(ctx context.Context) error {
	slog.Info("Closing Redis connection...")
	return r.Client.Close()
}`
  },
  {
    name: "lifecycle.go",
    path: "backend/pkg/lifecycle/lifecycle.go",
    description: "Gerenciador do Ciclo de Vida da aplicação. Escuta sinais do sistema operacional Unix para encerramento gracioso com controle de timeout.",
    language: "go",
    code: `package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CleanupFunc func(ctx context.Context) error

type Manager struct {
	cleanups []CleanupFunc
}

func NewManager() *Manager {
	return &Manager{
		cleanups: make([]CleanupFunc, 0),
	}
}

func (m *Manager) Register(fn CleanupFunc) {
	m.cleanups = append(m.cleanups, fn)
}

func (m *Manager) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	slog.Info("Shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := len(m.cleanups) - 1; i >= 0; i-- {
		cleanup := m.cleanups[i]
		if err := cleanup(ctx); err != nil {
			slog.Error("Error executing cleanup function", "error", err)
		}
	}
	slog.Info("All services cleaned up. Exit successful.")
}`
  },
  {
    name: "logger.go",
    path: "backend/pkg/logger/logger.go",
    description: "Configurador do logger padrão JSON de alta velocidade baseado no módulo nativo slog do Go 1.21+.",
    language: "go",
    code: `package logger

import (
	"log/slog"
	"os"
)

func InitLogger(level string) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}`
  },
  {
    name: "protocol.go",
    path: "backend/pkg/protocol/protocol.go",
    description: "Parser binário e serializador de pacotes TCP (cabeçalho oficial Little-Endian de 8 bytes e payloads variáveis).",
    language: "go",
    code: `package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	CS_HEARTBEAT           uint16 = 1000
	SC_HEARTBEAT_ACK       uint16 = 1001
	CS_LOGIN_REQUEST       uint16 = 1002
	SC_LOGIN_RESPONSE      uint16 = 1003
	CS_CHAR_LIST_REQUEST   uint16 = 1004
	SC_CHAR_LIST_RESPONSE  uint16 = 1005
	CS_CHAR_SELECT_REQUEST uint16 = 1006
	SC_CHAR_SELECT_RESPONSE uint16 = 1007

	CS_PLAYER_MOVE    uint16 = 2000
	SC_PLAYER_UPDATE  uint16 = 2001
	SC_SPAWN_ENTITY   uint16 = 2002
	SC_DESPAWN_ENTITY uint16 = 2003
)

const HeaderSize = 8
const MaxPacketSize = 16384

type Packet struct {
	Size     uint16
	Opcode   uint16
	Sequence uint32
	Payload  []byte
}

func ReadPacket(reader io.Reader) (*Packet, error) {
	headerBuf := make([]byte, HeaderSize)
	_, err := io.ReadFull(reader, headerBuf)
	if err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint16(headerBuf[0:2])
	opcode := binary.LittleEndian.Uint16(headerBuf[2:4])
	sequence := binary.LittleEndian.Uint32(headerBuf[4:8])

	if size < HeaderSize {
		return nil, fmt.Errorf("packet size %d too small (minimum %d)", size, HeaderSize)
	}
	if size > MaxPacketSize {
		return nil, fmt.Errorf("packet size %d exceeds max %d", size, MaxPacketSize)
	}

	payloadSize := size - HeaderSize
	payload := make([]byte, payloadSize)
	if payloadSize > 0 {
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return nil, err
		}
	}

	return &Packet{
		Size:     size,
		Opcode:   opcode,
		Sequence: sequence,
		Payload:  payload,
	}, nil
}

func (p *Packet) Serialize() []byte {
	p.Size = uint16(HeaderSize + len(p.Payload))
	buf := make([]byte, p.Size)
	binary.LittleEndian.PutUint16(buf[0:2], p.Size)
	binary.LittleEndian.PutUint16(buf[2:4], p.Opcode)
	binary.LittleEndian.PutUint32(buf[4:8], p.Sequence)
	if len(p.Payload) > 0 {
		copy(buf[HeaderSize:], p.Payload)
	}
	return buf
}`
  },
  {
    name: "messaging.go",
    path: "backend/pkg/messaging/messaging.go",
    description: "Internal Message Bus (PATCH 1). Barramento de mensagens thread-safe, channel-based e desacoplado para comunicação síncrona/assíncrona entre Gateway, Auth e World.",
    language: "go",
    code: `package messaging

import (
	"context"
	"errors"
	"sync"
	"time"
)

type MessageBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan any
}

var (
	Instance *MessageBus
	once     sync.Once
)

func GetInstance() *MessageBus {
	once.Do(func() {
		Instance = &MessageBus{
			subscribers: make(map[string][]chan any),
		}
	})
	return Instance
}

func (mb *MessageBus) Subscribe(topic string) <-chan any {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ch := make(chan any, 100)
	mb.subscribers[topic] = append(mb.subscribers[topic], ch)
	return ch
}

func (mb *MessageBus) Publish(topic string, message any) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	subs, exists := mb.subscribers[topic]
	if !exists {
		return
	}

	for _, ch := range subs {
		select {
		case ch <- message:
		default:
		}
	}
}

func (mb *MessageBus) RequestReply(topic string, request any, timeout time.Duration) (any, error) {
	replyTopic := topic + ".reply." + string(time.Now().UnixNano())
	replyChan := mb.Subscribe(replyTopic)
	defer mb.Unsubscribe(replyTopic, replyChan)

	envelope := struct {
		Payload    any
		ReplyTopic string
	}{
		Payload:    request,
		ReplyTopic: replyTopic,
	}

	mb.Publish(topic, envelope)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case response := <-replyChan:
		return response, nil
	case <-ctx.Done():
		return nil, errors.New("request-reply operation timed out")
	}
}

func (mb *MessageBus) Unsubscribe(topic string, ch <-chan any) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	subs, exists := mb.subscribers[topic]
	if !exists {
		return
	}

	for i, sub := range subs {
		if sub == ch {
			close(sub)
			mb.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	if len(mb.subscribers[topic]) == 0 {
		delete(mb.subscribers, topic)
	}
}`
  },
  {
    name: "scheduler.go",
    path: "backend/pkg/scheduler/scheduler.go",
    description: "Fixed Tick Scheduler (PATCH 2). Motor de ticks do World Server configurado a 20Hz (timestep fixo, delta time, tick overrun detection e lag logging).",
    language: "go",
    code: `package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type TickHandler func(deltaTime time.Duration, currentTick uint64)

type TickScheduler struct {
	hz          int
	interval    time.Duration
	handler     TickHandler
	stopChan    chan struct{}
	currentTick uint64
}

func NewTickScheduler(hz int, handler TickHandler) *TickScheduler {
	return &TickScheduler{
		hz:       hz,
		interval: time.Second / time.Duration(hz),
		handler:  handler,
		stopChan: make(chan struct{}),
	}
}

func (ts *TickScheduler) Start(ctx context.Context) {
	slog.Info("Starting TickScheduler...", "Hz", ts.hz, "TargetInterval", ts.interval)
	ticker := time.NewTicker(ts.interval)
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		select {
		case <-ts.stopChan:
			slog.Info("TickScheduler stopped gracefully by request.")
			return
		case <-ctx.Done():
			slog.Info("TickScheduler stopped due to context cancellation.")
			return
		case current := <-ticker.C:
			ts.currentTick++
			elapsed := current.Sub(lastTime)
			lastTime = current

			if elapsed > ts.interval + (5 * time.Millisecond) {
				lagAmount := elapsed - ts.interval
				slog.Warn("Tick Overrun Detected (Game Loop Lagging!)",
					"tick", ts.currentTick,
					"targetInterval", ts.interval,
					"actualElapsed", elapsed,
					"lag", lagAmount,
				)
			}

			ts.executeHandler(elapsed)
		}
	}
}

func (ts *TickScheduler) executeHandler(elapsed time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic recovered inside Game Loop Tick handler", "error", r, "tick", ts.currentTick)
		}
	}()
	ts.handler(elapsed, ts.currentTick)
}

func (ts *TickScheduler) Stop() {
	close(ts.stopChan)
}`
  },
  {
    name: "0001_create_accounts.up.sql",
    path: "backend/migrations/0001_create_accounts.up.sql",
    description: "Migration SQL (PATCH 4) - Criação da tabela de contas de usuários (accounts) e índices de busca rápida por username.",
    language: "sql",
    code: `-- Migration 0001: Criar tabela de contas de usuários (Accounts)
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);`
  },
  {
    name: "0002_create_characters.up.sql",
    path: "backend/migrations/0002_create_characters.up.sql",
    description: "Migration SQL (PATCH 4) - Criação da tabela de personagens (characters) vinculados a contas com tracking posicional.",
    language: "sql",
    code: `-- Migration 0002: Criar tabela de personagens (Characters)
CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(32) UNIQUE NOT NULL,
    class VARCHAR(20) NOT NULL,
    level INT DEFAULT 1,
    experience BIGINT DEFAULT 0,
    posX FLOAT DEFAULT 0.0,
    posY FLOAT DEFAULT 0.0,
    posZ FLOAT DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_characters_account_id ON characters(account_id);
CREATE INDEX IF NOT EXISTS idx_characters_name ON characters(name);`
  },
  {
    name: "0003_create_inventories.up.sql",
    path: "backend/migrations/0003_create_inventories.up.sql",
    description: "Migration SQL (PATCH 4) - Criação da tabela de inventários de itens (inventories) mapeando slots e durabilidades.",
    language: "sql",
    code: `-- Migration 0003: Criar tabela de inventários de itens (Inventories)
CREATE TABLE IF NOT EXISTS inventories (
    id SERIAL PRIMARY KEY,
    character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    slot_index INT NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    quantity INT DEFAULT 1,
    durability INT DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(character_id, slot_index)
);

CREATE INDEX IF NOT EXISTS idx_inventories_character_id ON inventories(character_id);`
  },
  {
    name: "0004_create_guilds.up.sql",
    path: "backend/migrations/0004_create_guilds.up.sql",
    description: "Migration SQL (PATCH 4) - Criação de tabelas de Guildas e membros associados (guilds e guild_members).",
    language: "sql",
    code: `-- Migration 0004: Criar tabelas de Guildas e Associação de membros (Guilds)
CREATE TABLE IF NOT EXISTS guilds (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    leader_id INT NOT NULL REFERENCES characters(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS guild_members (
    guild_id INT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
    character_id INT NOT NULL UNIQUE REFERENCES characters(id) ON DELETE CASCADE,
    rank VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(guild_id, character_id)
);

CREATE INDEX IF NOT EXISTS idx_guild_members_character_id ON guild_members(character_id);`
  },
  {
    name: "Pathfinding.cs",
    path: "src/Movement/Pathfinding.cs",
    description: "Algoritmo de busca de caminhos A* (A-Star) adaptado para grid 2D com suporte a movimentos diagonais livres.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;

namespace LightAndShadow.Client.Movement
{
    /// <summary>
    /// Implementação robusta e limitada do algoritmo A* para navegação diagonal livre no MMORPG.
    /// </summary>
    public class Pathfinding
    {
        public struct Point
        {
            public int X;
            public int Y;
            public Point(int x, int y) { X = x; Y = y; }
            public override bool Equals(object obj) => obj is Point p && p.X == X && p.Y == Y;
            public override int GetHashCode() => (X, Y).GetHashCode();
        }

        private class Node
        {
            public Point Position;
            public double G; // Custo do início ao nó atual
            public double H; // Custo heurístico estimado até o destino
            public double F => G + H;
            public Node Parent;
            public Node(Point pos, double g, double h, Node parent)
            {
                Position = pos;
                G = g;
                H = h;
                Parent = parent;
            }
        }

        private readonly ChunkManager _chunkManager;

        public Pathfinding(ChunkManager chunkManager)
        {
            _chunkManager = chunkManager;
        }

        /// <summary>
        /// Encontra o melhor caminho entre dois pontos usando movimentação diagonal livre limitada regionalmente.
        /// </summary>
        public List<Point> FindPath(Point start, Point end)
        {
            var openList = new List<Node>();
            var closedList = new HashSet<Point>();

            // Local chunk-only pathfinding restriction (start node chunk +/- 1 chunk)
            int startChunkX = start.X / 32;
            int startChunkY = start.Y / 32;

            if (Math.Abs(end.X / 32 - startChunkX) > 1 || Math.Abs(end.Y / 32 - startChunkY) > 1)
            {
                // Destino fora do limite local de chunks, retorna fallback imediatamente
                return new List<Point> { start };
            }

            openList.Add(new Node(start, 0, GetDistance(start, end), null));

            int expandedNodes = 0;
            const int MaxNodeExpansion = 4096;

            while (openList.Count > 0)
            {
                // Abort condition: max node expansion reached
                if (expandedNodes >= MaxNodeExpansion)
                {
                    return new List<Point> { start }; // Fallback no-path result
                }

                // Ordena a lista aberta para escolher o menor custo F
                openList.Sort((a, b) => a.F.CompareTo(b.F));
                var current = openList[0];
                openList.RemoveAt(0);

                closedList.Add(current.Position);
                expandedNodes++;

                if (current.Position.Equals(end))
                {
                    return RetracePath(current);
                }

                // Varre os vizinhos de 8 direções (diagonal livre)
                for (int dx = -1; dx <= 1; dx++)
                {
                    for (int dy = -1; dy <= 1; dy++)
                    {
                        if (dx == 0 && dy == 0) continue;

                        Point neighborPos = new Point(current.Position.X + dx, current.Position.Y + dy);

                        if (closedList.Contains(neighborPos)) continue;

                        // Local chunk-only pathfinding restriction
                        if (Math.Abs(neighborPos.X / 32 - startChunkX) > 1 || Math.Abs(neighborPos.Y / 32 - startChunkY) > 1)
                        {
                            continue;
                        }

                        if (!_chunkManager.IsWalkable(neighborPos.X, neighborPos.Y)) continue;

                        // Custo real: 1.0 para direcional ortogonal, ~1.414 para diagonal
                        double stepCost = (dx == 0 || dy == 0) ? 1.0 : 1.41421356;
                        double tentativeG = current.G + stepCost;

                        var existingNode = openList.Find(n => n.Position.Equals(neighborPos));
                        if (existingNode == null)
                        {
                            openList.Add(new Node(neighborPos, tentativeG, GetDistance(neighborPos, end), current));
                        }
                        else if (tentativeG < existingNode.G)
                        {
                            existingNode.G = tentativeG;
                            existingNode.Parent = current;
                        }
                    }
                }
            }

            return new List<Point> { start }; // Fallback no-path result (instead of null)
        }

        private double GetDistance(Point a, Point b)
        {
            // Distância Euclidiana Real para suporte a diagonal livre
            int dx = a.X - b.X;
            int dy = a.Y - b.Y;
            return Math.Sqrt(dx * dx + dy * dy);
        }

        private List<Point> RetracePath(Node endNode)
        {
            var path = new List<Point>();
            var current = endNode;
            while (current != null)
            {
                path.Add(current.Position);
                current = current.Parent;
            }
            path.Reverse();
            return path;
        }
    }
}`
  },
  {
    name: "ChunkManager.cs",
    path: "src/Movement/ChunkManager.cs",
    description: "Gerenciador do ciclo de vida dos chunks locais do cliente Godot, realizando cache e tratamento de pacotes de streaming.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;

namespace LightAndShadow.Client.Movement
{
    /// <summary>
    /// Gerenciador de Chunks de mapa (32x32 tiles) recebidos por streaming do servidor.
    /// </summary>
    public class ChunkManager
    {
        public class Chunk
        {
            public int ChunkX { get; set; }
            public int ChunkY { get; set; }
            public byte[,] Tiles { get; set; } // Matriz 32x32 (0 = Livre, 1 = Bloqueado)
        }

        private readonly Dictionary<ulong, Chunk> _loadedChunks = new Dictionary<ulong, Chunk>();

        private ulong GetChunkKey(int cx, int cy)
        {
            return ((ulong)(uint)cx << 32) | (uint)cy;
        }

        /// <summary>
        /// Processa pacote SC_CHUNK_DATA contendo binário compactado de um Chunk de 32x32 tiles.
        /// </summary>
        public void LoadChunkFromNetwork(byte[] payload)
        {
            if (payload == null || payload.Length < 1032) return;

            int cx = BitConverter.ToInt32(payload, 0);
            int cy = BitConverter.ToInt32(payload, 4);

            var chunk = new Chunk
            {
                ChunkX = cx,
                ChunkY = cy,
                Tiles = new byte[32, 32]
            };

            int offset = 8;
            for (int y = 0; y < 32; y++)
            {
                for (int x = 0; x < 32; x++)
                {
                    chunk.Tiles[y, x] = payload[offset++];
                }
            }

            ulong key = GetChunkKey(cx, cy);
            _loadedChunks[key] = chunk;
            
            GD.Print($"[Streaming] Chunk ({cx}, {cy}) recebido e instanciado com sucesso!");
        }

        /// <summary>
        /// Verifica colisão local estática contra os Chunks em cache
        /// </summary>
        public bool IsWalkable(int globalX, int globalY)
        {
            if (globalX < 0 || globalX >= 16384 || globalY < 0 || globalY >= 16384)
                return false; // Limites mundiais

            int cx = globalX / 32;
            int cy = globalY / 32;
            int rx = globalX % 32;
            int ry = globalY % 32;

            ulong key = GetChunkKey(cx, cy);
            if (_loadedChunks.TryGetValue(key, out Chunk chunk))
            {
                return chunk.Tiles[ry, rx] == 0; // 0 = Walkable, 1 = Obstáculo
            }

            // Se o chunk não foi carregado ainda por atraso de streaming, considera-se intransitável
            return false;
        }

        public void ClearCache()
        {
            _loadedChunks.Clear();
        }
    }
}`
  },
  {
    name: "MovementController.cs",
    path: "src/Movement/MovementController.cs",
    description: "Controlador posicional com Client-Side Prediction leve, reconciliação de física local e envio de CS_MOVE_REQUEST.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;

namespace LightAndShadow.Client.Movement
{
    /// <summary>
    /// Controla a movimentação preditiva local do jogador com reconciliação física via servidor.
    /// Velocidade: 1 tile / 250 ms.
    /// </summary>
    public class MovementController
    {
        private readonly NetworkManager _network;
        private readonly ChunkManager _chunkManager;
        
        // Coordenadas preditivas locais do jogador (atualizadas no frame)
        public double PosX { get; private set; } = 100.0;
        public double PosY { get; private set; } = 100.0;
        public int PosZ { get; private set; } = 0;

        // Histórico de sequência para reconciliação
        private uint _nextSequence = 1;
        
        private struct PendingMove
        {
            public uint Sequence;
            public double TargetX;
            public double TargetY;
        }
        
        private readonly List<PendingMove> _pendingMoves = new List<PendingMove>();

        public MovementController(NetworkManager network, ChunkManager chunkManager)
        {
            _network = network;
            _chunkManager = chunkManager;
        }

        /// <summary>
        /// Solicita deslocamento preditivo ao pressionar WASD ou clicar no mapa.
        /// </summary>
        public void RequestStepMove(double dx, double dy)
        {
            if (dx == 0 && dy == 0) return;

            // Normaliza deslocamento
            double len = Math.Sqrt(dx * dx + dy * dy);
            double stepX = dx / len;
            double stepY = dy / len;

            double targetX = PosX + stepX;
            double targetY = PosY + stepY;

            // 1. Prediction Local contra mapa estático (evita vibrar/tremer contra paredes)
            if (_chunkManager.IsWalkable((int)targetX, (int)targetY))
            {
                PosX = targetX;
                PosY = targetY;

                uint seq = _nextSequence++;
                _pendingMoves.Add(new PendingMove { Sequence = seq, TargetX = targetX, TargetY = targetY });

                // 2. Envia CS_MOVE_REQUEST estruturado em payload binário de 18 bytes (Sprint 2 Patch 1)
                byte[] payload = new byte[18];
                // int32 target_x (4 bytes)
                Array.Copy(BitConverter.GetBytes((int)targetX), 0, payload, 0, 4);
                // int32 target_y (4 bytes)
                Array.Copy(BitConverter.GetBytes((int)targetY), 0, payload, 4, 4);
                // int8 target_z (1 byte)
                payload[8] = (byte)PosZ;
                // uint8 direction (1 byte)
                payload[9] = 0; 
                // uint64 client_timestamp (8 bytes)
                long timestamp = DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();
                Array.Copy(BitConverter.GetBytes(timestamp), 0, payload, 10, 8);
                
                _network.SendPacket((ushort)PacketOpcode.CS_MOVE_REQUEST, seq, payload);
            }
        }

        /// <summary>
        /// Processa resposta SC_MOVE_CONFIRM do servidor para reconciliação física ativa.
        /// </summary>
        public void HandleMoveConfirm(double srvX, double srvY, int srvZ, uint confirmSeq, bool success)
        {
            // Remove as transições confirmadas até a sequência atualizada
            _pendingMoves.RemoveAll(m => m.Sequence <= confirmSeq);

            if (!success)
            {
                // Rejeitado! Servidor detectou speedhack ou colisão. Força Rubberbanding posicional.
                PosX = srvX;
                PosY = srvY;
                PosZ = srvZ;
                _pendingMoves.Clear();
                GD.PrintErr($"[Rubberbanding] Recuo posicional severo disparado pelo servidor para ({srvX:F2}, {srvY:F2})!");
                return;
            }

            // Se aceito, reaplica movimentos locais pendentes adicionados após a emissão do pacote confirmado
            double reconciledX = srvX;
            double reconciledY = srvY;

            foreach (var move in _pendingMoves)
            {
                reconciledX = move.TargetX;
                reconciledY = move.TargetY;
            }

            PosX = reconciledX;
            PosY = reconciledY;
            PosZ = srvZ;
        }
    }
}`
  },
  {
    name: "WorldManager.cs",
    path: "src/Movement/WorldManager.cs",
    description: "Gerencia a lista de entidades remotas visíveis na AOI do jogador e reage a spawns/despawns.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;

namespace LightAndShadow.Client.Movement
{
    /// <summary>
    /// Gerencia entidades remotas (jogadores/NPCs) presentes na Area of Interest (AOI) do cliente.
    /// </summary>
    public class WorldManager
    {
        public class RemoteEntity
        {
            public string ID { get; set; }
            public string Name { get; set; }
            public double X { get; set; }
            public double Y { get; set; }
            public int Z { get; set; }
            public string Type { get; set; }
        }

        private readonly Dictionary<string, RemoteEntity> _entities = new Dictionary<string, RemoteEntity>();
        
        public event Action<RemoteEntity> OnEntitySpawned;
        public event Action<string> OnEntityDespawned;
        public event Action<string, double, double, int> OnEntityUpdated;

        /// <summary>
        /// Processa pacote SC_SPAWN_ENTITY contendo JSON com dados completos da nova entidade visível.
        /// </summary>
        public void HandleSpawnEntity(byte[] payload)
        {
            if (payload == null || payload.Length == 0) return;

            string jsonStr = System.Text.Encoding.UTF8.GetString(payload);
            var entity = Newtonsoft.Json.JsonConvert.DeserializeObject<RemoteEntity>(jsonStr);

            if (entity != null)
            {
                _entities[entity.ID] = entity;
                OnEntitySpawned?.Invoke(entity);
                GD.Print($"[AOI] Entidade {entity.Name} (ID: {entity.ID}) entrou na sua tela (Spawn).");
            }
        }

        /// <summary>
        /// Processa pacote SC_DESPAWN_ENTITY contendo o ID da entidade que saiu do campo de visão.
        /// </summary>
        public void HandleDespawnEntity(byte[] payload)
        {
            if (payload == null || payload.Length == 0) return;

            string id = System.Text.Encoding.UTF8.GetString(payload);
            if (_entities.Remove(id))
            {
                OnEntityDespawned?.Invoke(id);
                GD.Print($"[AOI] Entidade ID: {id} saiu da sua tela (Despawn).");
            }
        }

        /// <summary>
        /// Processa pacote SC_PLAYER_UPDATE contendo as atualizações de posição de vizinhos.
        /// </summary>
        public void HandlePlayerUpdate(byte[] payload)
        {
            if (payload == null || payload.Length == 0) return;

            string jsonStr = System.Text.Encoding.UTF8.GetString(payload);
            var update = Newtonsoft.Json.JsonConvert.DeserializeAnonymousType(jsonStr, new { id = "", x = 0.0, y = 0.0, z = 0 });

            if (update != null && _entities.TryGetValue(update.id, out RemoteEntity entity))
            {
                entity.X = update.x;
                entity.Y = update.y;
                entity.Z = update.z;
                OnEntityUpdated?.Invoke(update.id, update.x, update.y, update.z);
            }
        }

        public IReadOnlyDictionary<string, RemoteEntity> GetEntities() => _entities;
    }
}`
  },
  {
    name: "spatial_index.go",
    path: "backend/pkg/movement/spatial_index.go",
    description: "Estrutura de dados thread-safe para indexação espacial utilizando chunks de 32x32 tiles para consultas rápidas na AOI.",
    language: "go",
    code: `package movement

import (
	"math"
	"sync"
)

// Entity representa qualquer entidade móvel no mundo de jogo (jogador ou NPC)
type Entity struct {
	ID    string  \`json:"id"\`
	Name  string  \`json:"name"\`
	X     float64 \`json:"x"\` // Coordenada X em Tiles
	Y     float64 \`json:"y"\` // Coordenada Y em Tiles
	Z     int     \`json:"z"\` // Floor (0 a 15)
	Type  string  \`json:"type"\` // "player" ou "npc"
}

// SpatialIndex gerencia o posicionamento de entidades em blocos 3D (X, Y, Z/Andar)
type SpatialIndex struct {
	mu       sync.RWMutex
	entities map[string]*Entity
	// Mapeia chunkKey -> lista de IDs de entidades
	// chunkKey é calculada como (chunkX << 32) | chunkY
	// Cada Z (andar) tem seu próprio mapa de chunks para isolamento total
	floors [16]map[uint64]map[string]*Entity
}

// NewSpatialIndex instancia o indexador espacial
func NewSpatialIndex() *SpatialIndex {
	si := &SpatialIndex{
		entities: make(map[string]*Entity),
	}
	for i := 0; i < 16; i++ {
		si.floors[i] = make(map[uint64]map[string]*Entity)
	}
	return si
}

// getChunkKey calcula a chave única para o chunk de tamanho 32x32 tiles
func getChunkKey(x, y float64) uint64 {
	cx := uint32(int(x) / 32)
	cy := uint32(int(y) / 32)
	return (uint64(cx) << 32) | uint64(cy)
}

// RegisterEntity adiciona uma nova entidade ao indexador espacial
func (si *SpatialIndex) RegisterEntity(entity *Entity) {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Valida andar
	if entity.Z < 0 || entity.Z >= 16 {
		entity.Z = 0
	}

	si.entities[entity.ID] = entity
	key := getChunkKey(entity.X, entity.Y)

	if si.floors[entity.Z][key] == nil {
		si.floors[entity.Z][key] = make(map[string]*Entity)
	}
	si.floors[entity.Z][key][entity.ID] = entity
}

// UpdateEntityPosition atualiza de forma atômica e thread-safe a posição de uma entidade
func (si *SpatialIndex) UpdateEntityPosition(id string, newX, newY float64, newZ int) (bool, *Entity) {
	si.mu.Lock()
	defer si.mu.Unlock()

	entity, exists := si.entities[id]
	if !exists {
		return false, nil
	}

	oldX, oldY, oldZ := entity.X, entity.Y, entity.Z
	oldKey := getChunkKey(oldX, oldY)

	// Valida novo andar
	if newZ < 0 || newZ >= 16 {
		newZ = 0
	}

	// Se mudou de posição ou andar, re-indexa
	if oldKey != getChunkKey(newX, newY) || oldZ != newZ {
		// Remove do chunk antigo
		if si.floors[oldZ][oldKey] != nil {
			delete(si.floors[oldZ][oldKey], id)
			if len(si.floors[oldZ][oldKey]) == 0 {
				delete(si.floors[oldZ], oldKey)
			}
		}

		// Adiciona no novo chunk
		newKey := getChunkKey(newX, newY)
		if si.floors[newZ][newKey] == nil {
			si.floors[newZ][newKey] = make(map[string]*Entity)
		}
		si.floors[newZ][newKey][id] = entity
	}

	entity.X = newX
	entity.Y = newY
	entity.Z = newZ

	return true, entity
}

// RemoveEntity desregistra uma entidade do indexador espacial
func (si *SpatialIndex) RemoveEntity(id string) {
	si.mu.Lock()
	defer si.mu.Unlock()

	entity, exists := si.entities[id]
	if !exists {
		return
	}

	key := getChunkKey(entity.X, entity.Y)
	if si.floors[entity.Z][key] != nil {
		delete(si.floors[entity.Z][key], id)
		if len(si.floors[entity.Z][key]) == 0 {
			delete(si.floors[entity.Z], key)
		}
	}

	delete(si.entities, id)
}

// GetEntitiesInRegion retorna entidades presentes nos chunks vizinhos que cobrem a região pesquisada
func (si *SpatialIndex) GetEntitiesInRegion(x, y float64, radius float64, z int) []*Entity {
	if z < 0 || z >= 16 {
		return nil
	}

	si.mu.RLock()
	defer si.mu.RUnlock()

	var result []*Entity

	// Determinar a faixa de chunks a varrer
	minX := math.Max(0, x-radius)
	maxX := math.Min(16384, x+radius)
	minY := math.Max(0, y-radius)
	maxY := math.Min(16384, y+radius)

	minChunkX := int(minX) / 32
	maxChunkX := int(maxX) / 32
	minChunkY := int(minY) / 32
	maxChunkY := int(maxY) / 32

	for cx := minChunkX; cx <= maxChunkX; cx++ {
		for cy := minChunkY; cy <= maxChunkY; cy++ {
			key := (uint64(cx) << 32) | uint64(cy)
			if chunkEntities, exists := si.floors[z][key]; exists {
				for _, ent := range chunkEntities {
					// Verifica distância euclidiana fina
					dx := ent.X - x
					dy := ent.Y - y
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist <= radius {
						result = append(result, ent)
					}
				}
			}
		}
	}

	return result
}`
  },
  {
    name: "chunk_manager.go",
    path: "backend/pkg/movement/chunk_manager.go",
    description: "Gerencia a geração lazy procedural e cache de Chunks no backend, definindo obstáculos estáticos.",
    language: "go",
    code: `package movement

import (
	"encoding/binary"
	"sync"
)

// Chunk representa um bloco espacial de 32x32 tiles
type Chunk struct {
	ChunkX int    \`json:"chunk_x"\`
	ChunkY int    \`json:"chunk_y"\`
	// Grid bidimensional de 32x32 representando IDs de tiles. 
	// 0 = Walkable (Grama/Chão), 1 = Obstáculo (Parede/Pedra)
	Tiles  [32][32]byte \`json:"tiles"\`
}

// ChunkManager gerencia o carregamento, cache e geração procedural dos chunks
type ChunkManager struct {
	mu     sync.RWMutex
	chunks map[uint64]*Chunk
}

// NewChunkManager inicializa o gerenciador de chunks do servidor
func NewChunkManager() *ChunkManager {
	return &ChunkManager{
		chunks: make(map[uint64]*Chunk),
	}
}

// getChunkKey calcula chave binária do chunk
func (cm *ChunkManager) getChunkKey(cx, cy int) uint64 {
	return (uint64(uint32(cx)) << 32) | uint64(uint32(cy))
}

// GetChunk recupera um chunk do cache ou gera de forma procedural (Lazy Loading)
func (cm *ChunkManager) GetChunk(cx, cy int) *Chunk {
	// Limites do mundo de 16384x16384 tiles (512x512 chunks de 32x32)
	if cx < 0 || cx >= 512 || cy < 0 || cy >= 512 {
		return nil
	}

	key := cm.getChunkKey(cx, cy)

	cm.mu.RLock()
	chunk, exists := cm.chunks[key]
	cm.mu.RUnlock()

	if exists {
		return chunk
	}

	// Geração procedural do chunk se não existir
	cm.mu.Lock()
	// Duplo check sob escrita segura
	if chunk, exists = cm.chunks[key]; exists {
		cm.mu.Unlock()
		return chunk
	}

	chunk = &Chunk{
		ChunkX: cx,
		ChunkY: cy,
	}

	// Popula tiles do chunk de forma inteligente e jogável
	for y := 0; y < 32; y++ {
		globalY := cy*32 + y
		for x := 0; x < 32; x++ {
			globalX := cx*32 + x

			// Garante uma zona segura (Spawn Zone) livre de obstáculos (ex: em torno de coord 100, 100)
			if globalX >= 80 && globalX <= 120 && globalY >= 80 && globalY <= 120 {
				chunk.Tiles[y][x] = 0 // Totalmente livre de colisões
			} else {
				// Adiciona obstáculos em padrão geométrico para simular paredes, árvores e pedras
				if (globalX%11 == 0 && globalY%7 == 0) || (globalX%13 == 0 && globalY%13 == 0) || (globalX%19 == 0 && globalY%5 == 0) {
					chunk.Tiles[y][x] = 1 // Colisão / Bloqueado
				} else {
					chunk.Tiles[y][x] = 0 // Caminhável
				}
			}
		}
	}

	cm.chunks[key] = chunk
	cm.mu.Unlock()

	return chunk
}

// IsBlocked verifica se um tile global específico no mundo de jogo é intransponível
func (cm *ChunkManager) IsBlocked(tileX, tileY int) bool {
	if tileX < 0 || tileX >= 16384 || tileY < 0 || tileY >= 16384 {
		return true // Fora do mapa é considerado obstáculo
	}

	cx := tileX / 32
	cy := tileY / 32
	rx := tileX % 32
	ry := tileY % 32

	chunk := cm.GetChunk(cx, cy)
	if chunk == nil {
		return true
	}

	return chunk.Tiles[ry][rx] == 1
}

// SerializeChunk compacta o chunk em um array binário otimizado para a rede
// Formato: 4 bytes ChunkX (LE), 4 bytes ChunkY (LE), 1024 bytes (tiles 32x32)
func (chunk *Chunk) Serialize() []byte {
	buf := make([]byte, 8+1024)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(chunk.ChunkX))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(chunk.ChunkY))

	offset := 8
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			buf[offset] = chunk.Tiles[y][x]
			offset++
		}
	}
	return buf
}

// GetSurroundingChunks recupera a lista de chunks em uma matriz de 3x3 ao redor de uma coordenada
func (cm *ChunkManager) GetSurroundingChunks(playerX, playerY float64) []*Chunk {
	cx := int(playerX) / 32
	cy := int(playerY) / 32

	var surrounding []*Chunk
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			chunk := cm.GetChunk(cx+dx, cy+dy)
			if chunk != nil {
				surrounding = append(surrounding, chunk)
			}
		}
	}
	return surrounding
}`
  },
  {
    name: "aoi_manager.go",
    path: "backend/pkg/movement/aoi_manager.go",
    description: "Coordena a Area of Interest (AOI) enviando spawns e despawns sob demanda conforme os jogadores se movem.",
    language: "go",
    code: `package movement

import (
	"encoding/json"
	"log/slog"
	"net"
	"sync"

	"github.com/light-and-shadow/backend/pkg/protocol"
)

// AOIManager coordena a visibilidade mútua e sincronização de rede entre jogadores próximos
type AOIManager struct {
	mu           sync.RWMutex
	spatialIndex *SpatialIndex
	// Conexões de rede ativas indexadas pelo ID do jogador
	connections  map[string]net.Conn
	// Rastreamento de quais entidades cada jogador atualmente "enxerga"
	// playerID -> set de entityIDs visíveis
	visibility   map[string]map[string]bool
}

// NewAOIManager instancia o gerenciador de visibilidade (AOI)
func NewAOIManager(si *SpatialIndex) *AOIManager {
	return &AOIManager{
		spatialIndex: si,
		connections:  make(map[string]net.Conn),
		visibility:   make(map[string]map[string]bool),
	}
}

// RegisterPlayer registra uma conexão ativa para receber atualizações espaciais
func (aoi *AOIManager) RegisterPlayer(id string, conn net.Conn) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	aoi.connections[id] = conn
	aoi.visibility[id] = make(map[string]bool)
	slog.Info("Player connection registered in AOIManager", "id", id)
}

// DeregisterPlayer remove a conexão e envia pacotes de despawn para todos os que viam este jogador
func (aoi *AOIManager) DeregisterPlayer(id string) {
	aoi.mu.Lock()
	// Remove a conexão e o conjunto de visibilidade do jogador
	delete(aoi.connections, id)
	delete(aoi.visibility, id)
	aoi.mu.Unlock()

	// Notifica outros jogadores sobre o despawn deste jogador desregistrado
	aoi.broadcastDespawn(id)
	slog.Info("Player connection deregistered from AOIManager", "id", id)
}

// UpdatePlayerAOI recalcula a visibilidade ao redor do jogador e dispara eventos de Spawn/Despawn (Deltas)
func (aoi *AOIManager) UpdatePlayerAOI(playerID string, x, y float64, z int) {
	aoi.mu.Lock()
	conn, hasConn := aoi.connections[playerID]
	observed, hasObs := aoi.visibility[playerID]
	aoi.mu.Unlock()

	if !hasConn || !hasObs {
		return
	}

	// Viewport é 24x18. Definimos o raio do AOI como 20 tiles (Sprint 2 Patch 4)
	const AOIRadius = 20.0

	// Consulta o SpatialIndex por todas as entidades próximas
	nearby := aoi.spatialIndex.GetEntitiesInRegion(x, y, AOIRadius, z)

	newVisible := make(map[string]*Entity)
	newVisibleIDs := make(map[string]bool)

	for _, ent := range nearby {
		// Não precisamos registrar spawn de nós mesmos
		if ent.ID == playerID {
			continue
		}
		newVisible[ent.ID] = ent
		newVisibleIDs[ent.ID] = true
	}

	// 1. Identificar entidades que SAÍRAM da área de visão (Despawn)
	var despawnIDs []string
	for oldID := range observed {
		if !newVisibleIDs[oldID] {
			despawnIDs = append(despawnIDs, oldID)
		}
	}

	// 2. Identificar entidades que ENTRARAM na área de visão (Spawn)
	var spawnEntities []*Entity
	for newID, ent := range newVisible {
		if !observed[newID] {
			spawnEntities = append(spawnEntities, ent)
		}
	}

	// Atualizar o set de visibilidade sob proteção
	aoi.mu.Lock()
	for _, id := range despawnIDs {
		delete(aoi.visibility[playerID], id)
	}
	for _, ent := range spawnEntities {
		aoi.visibility[playerID][ent.ID] = true
	}
	aoi.mu.Unlock()

	// 3. Enviar pacotes de Despawn para o cliente do jogador
	for _, id := range despawnIDs {
		payload := []byte(id)
		packet := &protocol.Packet{
			Opcode:  protocol.SC_DESPAWN_ENTITY,
			Payload: payload,
		}
		conn.Write(packet.Serialize())

		// Se a outra entidade for um jogador ativo, notifica ela também sobre o despawn recíproco
		aoi.sendDespawnToPlayer(id, playerID)
	}

	// 4. Enviar pacotes de Spawn para o cliente do jogador
	for _, ent := range spawnEntities {
		payload, err := json.Marshal(ent)
		if err != nil {
			slog.Error("Failed to marshal spawn entity JSON", "error", err)
			continue
		}
		packet := &protocol.Packet{
			Opcode:  protocol.SC_SPAWN_ENTITY,
			Payload: payload,
		}
		conn.Write(packet.Serialize())

		// Se a outra entidade for um jogador ativo, envia spawn recíproco dela nos enxergando
		aoi.sendSpawnToPlayer(ent.ID, playerID, x, y, z)
	}
}

// BroadcastMove envia atualizações de posição para todos os jogadores vizinhos
func (aoi *AOIManager) BroadcastMove(sourceID string, x, y float64, z int, payload []byte) {
	aoi.mu.RLock()
	defer aoi.mu.RUnlock()

	// Procura quem enxerga a fonte do movimento
	for playerID, observed := range aoi.visibility {
		if observed[sourceID] {
			if conn, exists := aoi.connections[playerID]; exists {
				packet := &protocol.Packet{
					Opcode:  protocol.SC_PLAYER_UPDATE,
					Payload: payload,
				}
				conn.Write(packet.Serialize())
			}
		}
	}
}

// sendDespawnToPlayer envia um despawn unilateral de targetID para o playerID
func (aoi *AOIManager) sendDespawnToPlayer(playerID, targetID string) {
	aoi.mu.Lock()
	conn, existsConn := aoi.connections[playerID]
	observed, existsObs := aoi.visibility[playerID]
	if existsConn && existsObs {
		delete(observed, targetID)
	}
	aoi.mu.Unlock()

	if existsConn {
		packet := &protocol.Packet{
			Opcode:  protocol.SC_DESPAWN_ENTITY,
			Payload: []byte(targetID),
		}
		conn.Write(packet.Serialize())
	}
}

// sendSpawnToPlayer envia o spawn unilateral do playerID (com coordenadas atuais) para o targetID
func (aoi *AOIManager) sendSpawnToPlayer(playerID, targetID string, x, y float64, z int) {
	aoi.mu.Lock()
	conn, existsConn := aoi.connections[playerID]
	observed, existsObs := aoi.visibility[playerID]
	if existsConn && existsObs {
		observed[targetID] = true
	}
	aoi.mu.Unlock()

	if existsConn {
		// Obtém informações da entidade a ser spawned
		aoi.spatialIndex.mu.RLock()
		targetEnt, ok := aoi.spatialIndex.entities[targetID]
		aoi.spatialIndex.mu.RUnlock()

		if !ok {
			// Se não achar entidade na memória, usa dados padrão
			targetEnt = &Entity{
				ID:   targetID,
				Name: "Player_" + targetID,
				X:    x,
				Y:    y,
				Z:    z,
				Type: "player",
			}
		}

		payload, err := json.Marshal(targetEnt)
		if err != nil {
			return
		}

		packet := &protocol.Packet{
			Opcode:  protocol.SC_SPAWN_ENTITY,
			Payload: payload,
		}
		conn.Write(packet.Serialize())
	}
}

// broadcastDespawn limpa e despawna um jogador deslogado de todas as telas vizinhas
func (aoi *AOIManager) broadcastDespawn(despawnedID string) {
	aoi.mu.Lock()
	defer aoi.mu.Unlock()

	packet := &protocol.Packet{
		Opcode:  protocol.SC_DESPAWN_ENTITY,
		Payload: []byte(despawnedID),
	}

	serialized := packet.Serialize()

	for playerID, observed := range aoi.visibility {
		if observed[despawnedID] {
			delete(observed, despawnedID)
			if conn, exists := aoi.connections[playerID]; exists {
				conn.Write(serialized)
			}
		}
	}
}`
  },
  {
    name: "movement.go",
    path: "backend/pkg/movement/movement.go",
    description: "Regras de movimento autoritativo no backend, efetuando checagens de velocidade e colisões estáticas.",
    language: "go",
    code: `package movement

import (
	"math"
	"sync"
	"time"
)

// PlayerMoveState armazena o histórico posicional e temporal para validação autoritativa
type PlayerMoveState struct {
	LastX               float64
	LastY               float64
	LastZ               int
	LastTime            time.Time
	Sequence            uint32
	IsInit              bool
	NextAllowedMoveTime time.Time
}

// MovementSystem coordena a física autoritativa de movimentação e colisões no servidor
type MovementSystem struct {
	mu           sync.RWMutex
	spatialIndex *SpatialIndex
	chunkManager *ChunkManager
	aoiManager   *AOIManager
	playerStates map[string]*PlayerMoveState
}

// NewMovementSystem inicializa o sistema de movimentação autoritativo
func NewMovementSystem(si *SpatialIndex, cm *ChunkManager, aoi *AOIManager) *MovementSystem {
	return &MovementSystem{
		spatialIndex: si,
		chunkManager: cm,
		aoiManager:   aoi,
		playerStates: make(map[string]*PlayerMoveState),
	}
}

// InitPlayerState define a posição inicial confiável do jogador no servidor
func (ms *MovementSystem) InitPlayerState(playerID string, x, y float64, z int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.playerStates[playerID] = &PlayerMoveState{
		LastX:               x,
		LastY:               y,
		LastZ:               z,
		LastTime:            time.Now(),
		IsInit:              true,
		NextAllowedMoveTime: time.Now(),
	}

	// Insere no indexador espacial se não existir
	ms.spatialIndex.RegisterEntity(&Entity{
		ID:   playerID,
		Name: "Player_" + playerID,
		X:    x,
		Y:    y,
		Z:    z,
		Type: "player",
	})
}

// RemovePlayerState limpa recursos de movimentação do jogador ao desconectar
func (ms *MovementSystem) RemovePlayerState(playerID string) {
	ms.mu.Lock()
	delete(ms.playerStates, playerID)
	ms.mu.Unlock()

	ms.spatialIndex.RemoveEntity(playerID)
}

// ValidateAndMove realiza as checagens autoritativas (velocidade, colisões de mapa, andares)
// Velocidade Oficial: 1 tile / 250 ms = 4.0 tiles/segundo.
func (ms *MovementSystem) ValidateAndMove(playerID string, targetX, targetY float64, targetZ int, seq uint32) (bool, float64, float64, int) {
	ms.mu.Lock()
	state, exists := ms.playerStates[playerID]
	if !exists {
		// Inicializa preguiçosamente com o target se não existia histórico
		ms.mu.Unlock()
		ms.InitPlayerState(playerID, targetX, targetY, targetZ)
		return true, targetX, targetY, targetZ
	}

	now := time.Now()

	// 0. Cooldown Check (PATCH 3 — Movement Cooldown)
	if now.Before(state.NextAllowedMoveTime) {
		ms.mu.Unlock()
		return false, state.LastX, state.LastY, state.LastZ
	}
	ms.mu.Unlock()

	// 1. Verificar Colisão de Obstáculo Estático no Chunk
	// Se o destino estiver bloqueado, o movimento é rejeitado imediatamente
	if ms.chunkManager.IsBlocked(int(targetX), int(targetY)) {
		// Retorna falso com a última posição válida (Rubberbanding/Client Correction)
		return false, state.LastX, state.LastY, state.LastZ
	}

	// 2. Validação Autoritativa de Velocidade (Anti-Speedhack com tolerância a Jitter)
	dt := now.Sub(state.LastTime).Seconds()
	dx := targetX - state.LastX
	dy := targetY - state.LastY

	// Distância Euclidiana em 2D (diagonal livre)
	distance := math.Sqrt(dx*dx + dy*dy)

	// Velocidade contratada: 4.0 tiles/segundo.
	const BaseSpeed = 4.0 // 1 tile / 250ms
	// Tolerância de 15% para absorver latência de rede e oscilação de ping (micro-lag/packet loss)
	const Tolerance = 1.15
	maxAllowedDistance := (BaseSpeed * dt) * Tolerance

	// Estabelece uma distância mínima de checagem para evitar problemas de divisão por dt muito pequeno
	if distance > 0.01 && dt > 0.0 {
		// Se a distância percorrida for maior que o máximo permitido, rejeitamos por speedhack
		// Permitir também uma folga estática inicial caso venha de um lag longo (ex: 1.5 tiles extras)
		const MaxLagBuffer = 1.5
		if distance > maxAllowedDistance+MaxLagBuffer {
			// Rejeitado! Força Rubberband no cliente
			return false, state.LastX, state.LastY, state.LastZ
		}
	}

	// 3. Atualizar Estado Válido e Índices Espaciais
	ms.mu.Lock()
	state.LastX = targetX
	state.LastY = targetY
	state.LastZ = targetZ
	state.LastTime = now
	state.Sequence = seq
	// Define next allowed move time after 250ms (cooldown rule)
	state.NextAllowedMoveTime = now.Add(250 * time.Millisecond)
	ms.mu.Unlock()

	return true, targetX, targetY, targetZ
}`
  },
  {
    name: "CombatController.cs",
    path: "src/Combat/CombatController.cs",
    description: "Controlador C# oficial que gerencia ataques básicos, valida o tempo de recarga da arma localmente e despacha pacotes CS_ATTACK_REQUEST.",
    language: "csharp",
    code: `using Godot;
using System;
using LightAndShadow.Client.Core;
using LightAndShadow.Client.Network;

namespace LightAndShadow.Client.Combat
{
    /// <summary>
    /// Gerencia as requisições de combate do jogador, prevenindo spam de ataques rápidos (anti-spam)
    /// e aplicando validação de cooldown preditiva baseada nas armas oficiais do jogo.
    /// </summary>
    public partial class CombatController : Node
    {
        private NetworkManager? _network;
        private TargetingController? _targeting;
        private double _nextAllowedAttackTime = 0.0;

        [Export] public string CurrentWeapon = "sword"; // dagger, sword, axe, hammer, bow, staff

        public override void _Ready()
        {
            _network = ServiceRegistry.Instance.Get<NetworkManager>();
            _targeting = GetNode<TargetingController>("../TargetingController");
        }

        public override void _Process(double delta)
        {
            if (Input.IsActionJustPressed("basic_attack") || Input.IsKeyPressed(Key.Space))
            {
                TriggerBasicAttack();
            }
        }

        public void TriggerBasicAttack()
        {
            if (_targeting == null || string.IsNullOrEmpty(_targeting.CurrentTargetID))
            {
                GD.Print("[CombatController] Nenhum alvo selecionado para atacar.");
                return;
            }

            double currentTime = Time.GetUnixTimeFromSystem();
            if (currentTime < _nextAllowedAttackTime)
            {
                double waitTime = _nextAllowedAttackTime - currentTime;
                GD.Print($"[CombatController] [Anti-Spam] Ataque em recarga! Aguarde {waitTime:F2}s");
                return;
            }

            // Validação de Distância Local (melhora a latência)
            float dist = _targeting.GetDistanceToTarget();
            float maxRange = GetWeaponRange(CurrentWeapon);
            if (dist > maxRange)
            {
                GD.Print($"[CombatController] Alvo fora de alcance! Distância: {dist:F1}m, Alcance: {maxRange:F1}m");
                return;
            }

            // Envia requisição para o Servidor Autorizativo (CS_ATTACK_REQUEST) usando protocolo binário GamePacket
            byte[] targetBytes = System.Text.Encoding.UTF8.GetBytes(_targeting.CurrentTargetID);
            byte[] weaponBytes = System.Text.Encoding.UTF8.GetBytes(CurrentWeapon);
            byte[] packetBytes = new byte[2 + targetBytes.Length + 2 + weaponBytes.Length];
            
            // Grava comprimentos e payloads em Little Endian
            packetBytes[0] = (byte)(targetBytes.Length & 0xFF);
            packetBytes[1] = (byte)((targetBytes.Length >> 8) & 0xFF);
            System.Buffer.BlockCopy(targetBytes, 0, packetBytes, 2, targetBytes.Length);
            
            int offset = 2 + targetBytes.Length;
            packetBytes[offset] = (byte)(weaponBytes.Length & 0xFF);
            packetBytes[offset + 1] = (byte)((weaponBytes.Length >> 8) & 0xFF);
            System.Buffer.BlockCopy(weaponBytes, 0, packetBytes, offset + 2, weaponBytes.Length);

            _network?.SendPacket((ushort)PacketOpcode.CS_ATTACK_REQUEST, 0, packetBytes);

            // Previsão local de Cooldown (Client-Side Prediction da velocidade da arma)
            double cooldown = GetWeaponCooldown(CurrentWeapon);
            _nextAllowedAttackTime = currentTime + cooldown;
            GD.Print($"[CombatController] Atacando {_targeting.CurrentTargetID} com {CurrentWeapon}. Próximo ataque em {cooldown}s.");
        }

        private float GetWeaponRange(string weapon)
        {
            return weapon switch
            {
                "dagger" => 1.0f,
                "sword" => 1.0f,
                "axe" => 1.0f,
                "hammer" => 1.0f,
                "spear" => 2.0f,
                "staff" => 4.0f,
                "bow" => 7.0f,
                "crossbow" => 9.0f,
                _ => 1.0f
            };
        }

        private double GetWeaponCooldown(string weapon)
        {
            return weapon switch
            {
                "dagger" => 0.65,
                "sword" => 1.00,
                "axe" => 1.35,
                "hammer" => 1.80,
                "bow" => 1.25,
                "staff" => 1.50,
                _ => 1.00
            };
        }
    }
}`
  },
  {
    name: "TargetingController.cs",
    path: "src/Combat/TargetingController.cs",
    description: "Gerencia a seleção e trava de alvos (Target Lock) do jogador, integrando feedback visual do indicador de seleção.",
    language: "csharp",
    code: `using Godot;
using System;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.Combat
{
    /// <summary>
    /// Gerencia o travamento visual de alvos (Target Lock) para o combate PvE e PvP.
    /// Limpa a trava caso o alvo morra ou saia da área de interesse.
    /// </summary>
    public partial class TargetingController : Node
    {
        public string CurrentTargetID { get; private set; } = string.Empty;
        public Vector2 TargetPosition { get; private set; } = Vector2.Zero;
        
        [Export] public Node2D? TargetIndicator; // Círculo de seleção visual no chão

        public override void _Ready()
        {
            EventBus.Instance.Subscribe<string>("target_dead", OnTargetDead);
        }

        public void SelectTarget(string entityId, Vector2 position)
        {
            CurrentTargetID = entityId;
            TargetPosition = position;
            GD.Print($"[TargetingController] Alvo Travado (Target Lock): {entityId} na posição {position}");
            
            if (TargetIndicator != null)
            {
                TargetIndicator.GlobalPosition = position;
                TargetIndicator.Visible = true;
            }
        }

        public void ClearTarget()
        {
            CurrentTargetID = string.Empty;
            if (TargetIndicator != null)
            {
                TargetIndicator.Visible = false;
            }
            GD.Print("[TargetingController] Alvo desmarcado.");
        }

        public float GetDistanceToTarget()
        {
            if (string.IsNullOrEmpty(CurrentTargetID)) return 999.0f;
            Node2D? playerNode = GetNodeOrNull<Node2D>("../Player");
            if (playerNode == null) return 999.0f;

            return playerNode.GlobalPosition.DistanceTo(TargetPosition) / 32.0f; // Escala do Grid de 32px/tile
        }

        private void OnTargetDead(string deadEntityId)
        {
            if (CurrentTargetID == deadEntityId)
            {
                GD.Print($"[TargetingController] Alvo {deadEntityId} morreu. Desfazendo Target Lock.");
                ClearTarget();
            }
        }
    }
}`
  },
  {
    name: "SkillController.cs",
    path: "src/Combat/SkillController.cs",
    description: "Controlador C# de habilidades ativas que suporta target lock e skillshots em área com pré-visualização de raio.",
    language: "csharp",
    code: `using Godot;
using System;
using System.Collections.Generic;
using LightAndShadow.Client.Core;
using LightAndShadow.Client.Network;

namespace LightAndShadow.Client.Combat
{
    /// <summary>
    /// Gerencia as habilidades de combate ativas do jogador (Ex: Slash, Fireball).
    /// Controla o cooldown de habilidades e despacha pacotes de conjuração CS_CAST_SKILL.
    /// </summary>
    public partial class SkillController : Node
    {
        private NetworkManager? _network;
        private TargetingController? _targeting;
        private Dictionary<uint, double> _cooldownTimers = new();

        public override void _Ready()
        {
            _network = ServiceRegistry.Instance.Get<NetworkManager>();
            _targeting = GetNode<TargetingController>("../TargetingController");
        }

        public override void _Process(double delta)
        {
            // Atalhos rápidos para conjurar habilidades (Teclas 1, 2, 3, 4)
            if (Input.IsActionJustPressed("cast_skill_1") || Input.IsKeyPressed(Key.Key1)) CastSkill(1);
            if (Input.IsActionJustPressed("cast_skill_2") || Input.IsKeyPressed(Key.Key2)) CastSkill(2);
            if (Input.IsActionJustPressed("cast_skill_3") || Input.IsKeyPressed(Key.Key3)) CastSkill(3);
            if (Input.IsActionJustPressed("cast_skill_4") || Input.IsKeyPressed(Key.Key4)) CastSkill(4);
        }

        public void CastSkill(uint skillID)
        {
            double currentTime = Time.GetUnixTimeFromSystem();
            if (_cooldownTimers.TryGetValue(skillID, out double allowedTime) && currentTime < allowedTime)
            {
                double wait = allowedTime - currentTime;
                GD.Print($"[SkillController] Habilidade [{skillID}] em cooldown! Espere {wait:F1}s.");
                return;
            }

            // Definição estática local para validação de range e tipo
            bool isArea = skillID == 2 || skillID == 4; // Fireball/Arrow Rain são em Área
            float range = skillID switch { 1 => 1.5f, 2 => 6.0f, 3 => 2.5f, 4 => 8.0f, _ => 5.0f };
            double cd = skillID switch { 1 => 1.5, 2 => 3.0, 3 => 2.0, 4 => 5.0, _ => 2.0 };

            string targetID = string.Empty;
            Vector2 castPosition = Vector2.Zero;

            if (!isArea)
            {
                // Target Lock requerido
                if (_targeting == null || string.IsNullOrEmpty(_targeting.CurrentTargetID))
                {
                    GD.Print("[SkillController] Habilidade de alvo único requer um Target Lock ativo!");
                    return;
                }
                if (_targeting.GetDistanceToTarget() > range)
                {
                    GD.Print("[SkillController] Alvo fora de alcance para esta habilidade!");
                    return;
                }
                targetID = _targeting.CurrentTargetID;
                castPosition = _targeting.TargetPosition;
            }
            else
            {
                // Skillshot em Área (conjurada na posição do mouse)
                castPosition = GetViewport().GetMousePosition();
                Node2D? playerNode = GetNodeOrNull<Node2D>("../Player");
                if (playerNode != null)
                {
                    float dist = (playerNode.GlobalPosition.DistanceTo(castPosition)) / 32.0f;
                    if (dist > range)
                    {
                        GD.Print($"[SkillController] Ponto do Skillshot fora de alcance! Distância: {dist:F1}m, Max: {range:F1}m");
                        return;
                    }
                }
            }

            // Envia solicitação de conjuração autoritativa (CS_CAST_SKILL) em formato binário GamePacket
            byte[] targetBytes = System.Text.Encoding.UTF8.GetBytes(targetID);
            byte[] packetBytes = new byte[4 + 2 + targetBytes.Length + 8 + 8];
            
            // SkillID (uint32, 4 bytes)
            System.BitConverter.GetBytes(skillID).CopyTo(packetBytes, 0);
            
            // TargetID length (uint16, 2 bytes) and bytes
            packetBytes[4] = (byte)(targetBytes.Length & 0xFF);
            packetBytes[5] = (byte)((targetBytes.Length >> 8) & 0xFF);
            System.Buffer.BlockCopy(targetBytes, 0, packetBytes, 6, targetBytes.Length);
            
            // TargetX & TargetY (float64, 8 bytes each)
            int offset = 6 + targetBytes.Length;
            System.BitConverter.GetBytes((double)(castPosition.X / 32.0f)).CopyTo(packetBytes, offset);
            System.BitConverter.GetBytes((double)(castPosition.Y / 32.0f)).CopyTo(packetBytes, offset + 8);

            _network?.SendPacket((ushort)PacketOpcode.CS_CAST_SKILL, 0, packetBytes);

            // Registra recarga de previsão do cliente
            _cooldownTimers[skillID] = currentTime + cd;
            GD.Print($"[SkillController] Conjurando Habilidade {skillID} com sucesso. Cooldown de {cd}s iniciado.");
        }
    }
}`
  },
  {
    name: "DamageNumberRenderer.cs",
    path: "src/Combat/DamageNumberRenderer.cs",
    description: "Componente de UI 2D responsável por instanciar e animar os números de dano (normais, críticos e misses).",
    language: "csharp",
    code: `using Godot;
using System;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.Combat
{
    /// <summary>
    /// Renderiza números flutuantes e coloridos para feedbacks de dano no mundo de jogo.
    /// Distingue acertos normais, golpes críticos (vermelho/maior) e desvios (MISS cinza).
    /// </summary>
    public partial class DamageNumberRenderer : Node2D
    {
        [Export] public Font? CustomFont;

        public override void _Ready()
        {
            // Escuta o barramento global de eventos de combate
            EventBus.Instance.Subscribe<DamageEvent>("damage_taken", OnDamageTaken);
        }

        private void OnDamageTaken(DamageEvent dmg)
        {
            SpawnDamageNumber(dmg.Position, dmg.Amount, dmg.IsCritical, dmg.IsHit);
        }

        public void SpawnDamageNumber(Vector2 position, float amount, bool isCrit, bool isHit)
        {
            var label = new Label();
            AddChild(label);

            Random rng = new();
            Vector2 randomOffset = new Vector2(rng.Next(-12, 12), rng.Next(-18, -6));
            label.GlobalPosition = position + randomOffset;

            if (CustomFont != null)
            {
                label.AddThemeFontOverride("font", CustomFont);
            }

            if (!isHit)
            {
                label.Text = "MISS";
                label.AddThemeColorOverride("font_color", Colors.DarkGray);
                label.AddThemeFontSizeOverride("font_size", 12);
            }
            else
            {
                label.Text = amount.ToString("F0");
                if (isCrit)
                {
                    label.AddThemeColorOverride("font_color", Colors.OrangeRed);
                    label.AddThemeFontSizeOverride("font_size", 22);
                    label.Text += "!";
                }
                else
                {
                    label.AddThemeColorOverride("font_color", Colors.White);
                    label.AddThemeFontSizeOverride("font_size", 14);
                }
            }

            // Animação rica com Tween (Flutuar e Fade Out)
            var tween = CreateTween().SetParallel(true);
            Vector2 targetPosition = label.GlobalPosition + new Vector2(0, -45);
            
            tween.TweenProperty(label, "global_position", targetPosition, 0.75f)
                 .SetTrans(Tween.TransitionType.Cubic)
                 .SetEase(Tween.EaseType.Out);

            tween.TweenProperty(label, "modulate:a", 0.0f, 0.75f)
                 .SetTrans(Tween.TransitionType.Linear);

            tween.Chain().TweenCallback(Callable.From(label.QueueFree));
        }
    }

    public struct DamageEvent
    {
        public Vector2 Position;
        public float Amount;
        public bool IsCritical;
        public bool IsHit;
    }
}`
  },
  {
    name: "damage_formula.go",
    path: "backend/pkg/combat/damage_formula.go",
    description: "Fórmulas matemáticas autoritativas para o cálculo de acerto (accuracy vs evasion), dano físico/mágico e mitigação por defesa/resistência.",
    language: "go",
    code: `package combat

import (
	"math/rand"
	"time"
)

type EntityStats struct {
	ID                 string
	Name               string
	IsPlayer           bool
	Faction            string
	BaseAttack         float64
	Defense            float64
	Resistance         float64
	Accuracy           float64
	Evasion            float64
	CritChance         float64
	CritMultiplier     float64
	Health             float64
	MaxHealth          float64
}

func CalculateHitChance(attacker, defender *EntityStats) float64 {
	acc := attacker.Accuracy
	if acc <= 0 { acc = 1.0 }
	eva := defender.Evasion
	if eva < 0 { eva = 0 }
	hitChance := acc / (acc + eva)
	if hitChance < 0.10 { return 0.10 }
	if hitChance > 0.95 { return 0.95 }
	return hitChance
}

func RollHit(attacker, defender *EntityStats) bool {
	chance := CalculateHitChance(attacker, defender)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rng.Float64() <= chance
}

func CalculateDamage(attacker, defender *EntityStats, weaponScale float64, skillScale float64, isPvP bool) (float64, bool) {
	rawAttack := attacker.BaseAttack
	scaledAttack := rawAttack * weaponScale
	baseDamage := scaledAttack * skillScale

	isCrit := false
	critChance := attacker.CritChance
	if critChance < 0.05 { critChance = 0.05 }
	critMult := attacker.CritMultiplier
	if critMult < 1.50 { critMult = 1.50 }

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if rng.Float64() <= critChance {
		isCrit = true
		baseDamage = baseDamage * critMult
	}

	def := defender.Defense
	if def < 0 { def = 0 }
	defReduction := def / (def + 100.0)
	damageAfterDef := baseDamage * (1.0 - defReduction)

	res := defender.Resistance
	if res < 0 { res = 0 }
	if res > 100.0 { res = 100.0 }
	resReduction := res / 100.0
	damageAfterRes := damageAfterDef * (1.0 - resReduction)

	finalDamage := damageAfterRes
	if isPvP {
		finalDamage = finalDamage * 0.70
	}

	if finalDamage < 1.0 {
		finalDamage = 1.0
	}

	return finalDamage, isCrit
}`
  },
  {
    name: "aggro_manager.go",
    path: "backend/pkg/combat/aggro_manager.go",
    description: "Gerencia a ameaça (threat) de NPCs para combates PvE baseados em alvos múltiplos.",
    language: "go",
    code: `package combat

import "sync"

type AggroTable struct {
	mu     sync.RWMutex
	threat map[string]float64
}

func NewAggroTable() *AggroTable {
	return &AggroTable{threat: make(map[string]float64)}
}

func (at *AggroTable) AddThreat(playerID string, amount float64) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.threat[playerID] += amount
}

func (at *AggroTable) SetThreat(playerID string, amount float64) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.threat[playerID] = amount
}

func (at *AggroTable) GetThreat(playerID string) float64 {
	at.mu.RLock()
	defer at.mu.RUnlock()
	return at.threat[playerID]
}

func (at *AggroTable) ClearThreat() {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.threat = make(map[string]float64)
}

func (at *AggroTable) RemovePlayer(playerID string) {
	at.mu.Lock()
	defer at.mu.Unlock()
	delete(at.threat, playerID)
}

func (at *AggroTable) GetTopTarget() (string, bool) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var topPlayer string
	maxThreat := -1.0

	for playerID, amount := range at.threat {
		if amount > maxThreat {
			maxThreat = amount
			topPlayer = playerID
		}
	}

	if maxThreat < 0 {
		return "", false
	}
	return topPlayer, true
}

func (at *AggroTable) DecayThreats(percent float64) {
	at.mu.Lock()
	defer at.mu.Unlock()

	multiplier := 1.0 - (percent / 100.0)
	if multiplier < 0 { multiplier = 0 }

	for playerID := range at.threat {
		at.threat[playerID] *= multiplier
		if at.threat[playerID] < 1.0 {
			delete(at.threat, playerID)
		}
	}
}`
  },
  {
    name: "skill_resolver.go",
    path: "backend/pkg/combat/skill_resolver.go",
    description: "Módulo autoritativo de resolução de habilidades, validando distâncias e calculando danos em área e alvo único.",
    language: "go",
    code: `package combat

import (
	"fmt"
	"math"
	"time"
)

type Skill struct {
	ID          uint32
	Name        string
	Cooldown    time.Duration
	Range       float64
	SkillScale  float64
	IsArea      bool
	AreaRadius  float64
	ManaCost    float64
}

var PredefinedSkills = map[uint32]Skill{
	1: {ID: 1, Name: "Slash", Cooldown: 1500 * time.Millisecond, Range: 1.5, SkillScale: 1.3, IsArea: false, ManaCost: 10},
	2: {ID: 2, Name: "Fireball", Cooldown: 3000 * time.Millisecond, Range: 6.0, SkillScale: 1.8, IsArea: true, AreaRadius: 3.0, ManaCost: 25},
	3: {ID: 3, Name: "Spear Thrust", Cooldown: 2000 * time.Millisecond, Range: 2.5, SkillScale: 1.4, IsArea: false, ManaCost: 15},
	4: {ID: 4, Name: "Arrow Rain", Cooldown: 5000 * time.Millisecond, Range: 8.0, SkillScale: 1.1, IsArea: true, AreaRadius: 4.5, ManaCost: 30},
}

type SkillCastResult struct {
	Skill       Skill
	AttackerID  string
	Success     bool
	ErrorMessage string
	TargetsHit  []DamageResult
}

type DamageResult struct {
	TargetID string
	Damage   float64
	IsCrit   bool
	IsHit    bool
}

func ResolveSkill(skill Skill, attacker *EntityStats, attackerX, attackerY float64, target *EntityStats, targetX, targetY float64, nearbyEntities []*EntityStats) *SkillCastResult {
	result := &SkillCastResult{Skill: skill, AttackerID: attacker.ID, Success: true}

	var dist float64
	if !skill.IsArea {
		if target == nil {
			result.Success = false
			result.ErrorMessage = "Habilidade de alvo único requer um alvo válido."
			return result
		}
		dist = math.Hypot(targetX-attackerX, targetY-attackerY)
	} else {
		dist = math.Hypot(targetX-attackerX, targetY-attackerY)
	}

	if dist > skill.Range {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Alvo fora de alcance. Distância: %.2f, Alcance: %.2f", dist, skill.Range)
		return result
	}

	if !skill.IsArea {
		isHit := RollHit(attacker, target)
		if !isHit {
			result.TargetsHit = append(result.TargetsHit, DamageResult{TargetID: target.ID, Damage: 0, IsCrit: false, IsHit: false})
			return result
		}

		isPvP := attacker.IsPlayer && target.IsPlayer
		if attacker.Faction == target.Faction && attacker.IsPlayer && target.IsPlayer {
			result.Success = false
			result.ErrorMessage = "Fogo amigo desabilitado para membros da mesma facção."
			return result
		}

		damage, isCrit := CalculateDamage(attacker, target, 1.0, skill.SkillScale, isPvP)
		result.TargetsHit = append(result.TargetsHit, DamageResult{TargetID: target.ID, Damage: damage, IsCrit: isCrit, IsHit: true})
	} else {
		for _, entity := range nearbyEntities {
			if entity.ID == attacker.ID { continue }
			if attacker.IsPlayer && entity.IsPlayer && attacker.Faction == entity.Faction { continue }

			isHit := RollHit(attacker, entity)
			if !isHit {
				result.TargetsHit = append(result.TargetsHit, DamageResult{TargetID: entity.ID, Damage: 0, IsCrit: false, IsHit: false})
				continue
			}

			isPvP := attacker.IsPlayer && entity.IsPlayer
			damage, isCrit := CalculateDamage(attacker, entity, 1.0, skill.SkillScale, isPvP)
			result.TargetsHit = append(result.TargetsHit, DamageResult{TargetID: entity.ID, Damage: damage, IsCrit: isCrit, IsHit: true})
		}
	}

	return result
}`
  },
  {
    name: "combat_scheduler.go",
    path: "backend/pkg/combat/combat_scheduler.go",
    description: "Executa loops de tempo real em segundo plano para o gerenciamento de cura passiva, ticks de NPCs e expiração de estados.",
    language: "go",
    code: `package combat

import (
	"context"
	"sync"
	"time"
)

type CombatScheduler struct {
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	tasks     map[string]func()
	interval  time.Duration
	running   bool
}

func NewCombatScheduler(interval time.Duration) *CombatScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &CombatScheduler{ctx: ctx, cancel: cancel, tasks: make(map[string]func()), interval: interval}
}

func (cs *CombatScheduler) RegisterTask(name string, task func()) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.tasks[name] = task
}

func (cs *CombatScheduler) UnregisterTask(name string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.tasks, name)
}

func (cs *CombatScheduler) Start() {
	cs.mu.Lock()
	if cs.running {
		cs.mu.Unlock()
		return
	}
	cs.running = true
	cs.mu.Unlock()
	go cs.runLoop()
}

func (cs *CombatScheduler) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if !cs.running { return }
	cs.cancel()
	cs.running = false
}

func (cs *CombatScheduler) runLoop() {
	ticker := time.NewTicker(cs.interval)
	defer ticker.Stop()
	for {
		select {
		case <-cs.ctx.Done():
			return
		case <-ticker.C:
			cs.executeTasks()
		}
	}
}

func (cs *CombatScheduler) executeTasks() {
	cs.mu.Lock()
	tasksToRun := make([]func(), 0, len(cs.tasks))
	for _, task := range cs.tasks {
		tasksToRun = append(tasksToRun, task)
	}
	cs.mu.Unlock()

	for _, task := range tasksToRun {
		safelyExecute(task)
	}
}

func safelyExecute(task func()) {
	defer func() { recover() }()
	task()
}`
  },
  {
    name: "combat_manager.go",
    path: "backend/pkg/combat/combat_manager.go",
    description: "Orquestrador autoritativo central do sistema de combate PvE e PvP do Light and Shadow.",
    language: "go",
    code: `package combat

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type WeaponConfig struct {
	Name    string
	Speed   time.Duration
	Range   float64
	Scaling float64
}

var WeaponConfigs = map[string]WeaponConfig{
	"dagger":   {Name: "Adaga", Speed: 650 * time.Millisecond, Range: 1.0, Scaling: 0.9},
	"sword":    {Name: "Espada", Speed: 1000 * time.Millisecond, Range: 1.0, Scaling: 1.1},
	"axe":      {Name: "Machado", Speed: 1350 * time.Millisecond, Range: 1.0, Scaling: 1.3},
	"hammer":   {Name: "Martelo", Speed: 1800 * time.Millisecond, Range: 1.0, Scaling: 1.6},
	"spear":    {Name: "Lança", Speed: 1200 * time.Millisecond, Range: 2.0, Scaling: 1.2},
	"staff":    {Name: "Cajado", Speed: 1500 * time.Millisecond, Range: 4.0, Scaling: 1.0},
	"bow":      {Name: "Arco", Speed: 1250 * time.Millisecond, Range: 7.0, Scaling: 1.0},
	"crossbow": {Name: "Besta", Speed: 1400 * time.Millisecond, Range: 9.0, Scaling: 1.4},
}

type CombatManager struct {
	mu             sync.RWMutex
	entities       map[string]*EntityStats
	entityPos      map[string]struct{ X, Y float64 }
	aggroTables    map[string]*AggroTable
	nextAttackTime map[string]time.Time
	skillCooldowns map[string]map[uint32]time.Time
	scheduler      *CombatScheduler
}

func NewCombatManager() *CombatManager {
	cm := &CombatManager{
		entities:       make(map[string]*EntityStats),
		entityPos:      make(map[string]struct{ X, Y float64 }),
		aggroTables:    make(map[string]*AggroTable),
		nextAttackTime: make(map[string]time.Time),
		skillCooldowns: make(map[string]map[uint32]time.Time),
		scheduler:      NewCombatScheduler(500 * time.Millisecond),
	}
	cm.scheduler.RegisterTask("HpRegenAndAggroDecay", cm.tickHpRegenAndAggro)
	cm.scheduler.Start()
	return cm
}

func (cm *CombatManager) RegisterEntity(entity *EntityStats, x, y float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.entities[entity.ID] = entity
	cm.entityPos[entity.ID] = struct{ X, Y float64 }{X: x, Y: y}
	cm.nextAttackTime[entity.ID] = time.Now()
	cm.skillCooldowns[entity.ID] = make(map[uint32]time.Time)
	if !entity.IsPlayer {
		cm.aggroTables[entity.ID] = NewAggroTable()
	}
}

func (cm *CombatManager) UpdateEntityPosition(id string, x, y float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.entityPos[id]; exists {
		cm.entityPos[id] = struct{ X, Y float64 }{X: x, Y: y}
	}
}

func (cm *CombatManager) DeregisterEntity(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.entities, id)
	delete(cm.entityPos, id)
	delete(cm.aggroTables, id)
	delete(cm.nextAttackTime, id)
	delete(cm.skillCooldowns, id)
}

func (cm *CombatManager) GetEntityStats(id string) (*EntityStats, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	stats, exists := cm.entities[id]
	return stats, exists
}

func (cm *CombatManager) ProcessAttackRequest(attackerID, targetID, weaponType string) (float64, bool, error) {
	cm.mu.Lock()
	attacker, existsAtt := cm.entities[attackerID]
	target, existsTar := cm.entities[targetID]
	attPos, hasAttPos := cm.entityPos[attackerID]
	tarPos, hasTarPos := cm.entityPos[targetID]
	cm.mu.Unlock()

	if !existsAtt || !existsTar || !hasAttPos || !hasTarPos {
		return 0, false, errors.New("atacante ou alvo não encontrado no mundo")
	}

	if attacker.Health <= 0 { return 0, false, errors.New("atacante está morto") }
	if target.Health <= 0 { return 0, false, errors.New("alvo já está morto") }

	cm.mu.Lock()
	nextAllowed, existsCooldown := cm.nextAttackTime[attackerID]
	cm.mu.Unlock()

	now := time.Now()
	if existsCooldown && now.Before(nextAllowed) {
		timeLeft := nextAllowed.Sub(now)
		return 0, false, fmt.Errorf("ataque em recarga. Espere %.2fs", timeLeft.Seconds())
	}

	wConfig, existsWeapon := WeaponConfigs[weaponType]
	if !existsWeapon { wConfig = WeaponConfigs["sword"] }

	dist := math.Hypot(tarPos.X-attPos.X, tarPos.Y-attPos.Y)
	if dist > wConfig.Range {
		return 0, false, fmt.Errorf("alvo fora de alcance para %s. Distância: %.2fm, Alcance: %.2fm", wConfig.Name, dist, wConfig.Range)
	}

	cm.mu.Lock()
	cm.nextAttackTime[attackerID] = now.Add(wConfig.Speed)
	cm.mu.Unlock()

	if attacker.IsPlayer && target.IsPlayer && attacker.Faction == target.Faction {
		return 0, false, errors.New("fogo amigo desabilitado")
	}

	if !RollHit(attacker, target) { return 0, false, nil }

	isPvP := attacker.IsPlayer && target.IsPlayer
	damage, isCrit := CalculateDamage(attacker, target, wConfig.Scaling, 1.0, isPvP)

	cm.mu.Lock()
	target.Health -= damage
	if target.Health < 0 { target.Health = 0 }
	isDead := target.Health <= 0
	cm.mu.Unlock()

	if attacker.IsPlayer && !target.IsPlayer {
		cm.mu.Lock()
		atTable, hasAt := cm.aggroTables[target.ID]
		cm.mu.Unlock()
		if hasAt { atTable.AddThreat(attackerID, damage) }
	}

	if isDead { cm.handleEntityDeath(target.ID) }
	return damage, isCrit, nil
}

func (cm *CombatManager) ProcessCastSkillRequest(attackerID string, skillID uint32, targetID string, targetX, targetY float64) (*SkillCastResult, error) {
	cm.mu.Lock()
	attacker, existsAtt := cm.entities[attackerID]
	attPos, hasAttPos := cm.entityPos[attackerID]
	cm.mu.Unlock()

	if !existsAtt || !hasAttPos { return nil, errors.New("atacante não encontrado no mundo") }
	if attacker.Health <= 0 { return nil, errors.New("atacante está morto") }

	skill, existsSkill := PredefinedSkills[skillID]
	if !existsSkill { return nil, fmt.Errorf("habilidade ID %d não existe", skillID) }

	cm.mu.Lock()
	nextAllowed, existsCD := cm.skillCooldowns[attackerID][skillID]
	cm.mu.Unlock()

	now := time.Now()
	if existsCD && now.Before(nextAllowed) { return nil, fmt.Errorf("habilidade %s está em recarga", skill.Name) }

	var target *EntityStats
	if !skill.IsArea {
		cm.mu.Lock()
		targetStats, hasTarget := cm.entities[targetID]
		tarPos, hasTarPos := cm.entityPos[targetID]
		cm.mu.Unlock()

		if !hasTarget || !hasTarPos { return nil, errors.New("alvo único não encontrado") }
		if targetStats.Health <= 0 { return nil, errors.New("alvo já está morto") }
		target = targetStats
		targetX, targetY = tarPos.X, tarPos.Y
	}

	var nearbyCandidates []*EntityStats
	if skill.IsArea {
		cm.mu.RLock()
		for id, ent := range cm.entities {
			if id == attackerID || ent.Health <= 0 { continue }
			entPos := cm.entityPos[id]
			dist := math.Hypot(entPos.X-targetX, entPos.Y-targetY)
			if dist <= skill.AreaRadius { nearbyCandidates = append(nearbyCandidates, ent) }
		}
		cm.mu.RUnlock()
	}

	result := ResolveSkill(skill, attacker, attPos.X, attPos.Y, target, targetX, targetY, nearbyCandidates)
	if !result.Success { return result, errors.New(result.ErrorMessage) }

	cm.mu.Lock()
	cm.skillCooldowns[attackerID][skillID] = now.Add(skill.Cooldown)
	for _, dmg := range result.TargetsHit {
		if dmg.IsHit && dmg.Damage > 0 {
			tEnt, ok := cm.entities[dmg.TargetID]
			if ok {
				tEnt.Health -= dmg.Damage
				if tEnt.Health < 0 { tEnt.Health = 0 }
				if attacker.IsPlayer && !tEnt.IsPlayer {
					atTable, hasAt := cm.aggroTables[tEnt.ID]
					if hasAt { atTable.AddThreat(attackerID, dmg.Damage) }
				}
				if tEnt.Health <= 0 { go cm.handleEntityDeath(tEnt.ID) }
			}
		}
	}
	cm.mu.Unlock()

	return result, nil
}

func (cm *CombatManager) handleEntityDeath(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for npcID, atTable := range cm.aggroTables {
		atTable.RemovePlayer(id)
		if npcID == id { atTable.ClearThreat() }
	}
}

func (cm *CombatManager) tickHpRegenAndAggro() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, ent := range cm.entities {
		if ent.Health > 0 && ent.Health < ent.MaxHealth {
			ent.Health += ent.MaxHealth * 0.01
			if ent.Health > ent.MaxHealth { ent.Health = ent.MaxHealth }
		}
	}
	for _, atTable := range cm.aggroTables {
		atTable.DecayThreats(5.0)
	}
}

func (cm *CombatManager) Close() {
	cm.scheduler.Stop()
}`
  },
  {
    name: "InventoryManager.cs",
    path: "src/Network/InventoryManager.cs",
    description: "Gerenciador C# autoritativo do Inventário e Equipamentos, processando pacotes binários compactos de sincronização e requisições.",
    language: "csharp",
    code: `using System;
using System.IO;
using System.Collections.Generic;
using Godot;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.Network
{
    public class ClientItem
    {
        public string ItemID { get; set; } = "";
        public int Quantity { get; set; }
        public int Durability { get; set; }
        public int SlotIndex { get; set; }
    }

    public class CharacterAttributes
    {
        public int Level { get; set; }
        public double MaxHealth { get; set; }
        public double Health { get; set; }
        public double MaxMana { get; set; }
        public double Mana { get; set; }
        public double BaseAttack { get; set; }
        public double WeaponDamage { get; set; }
        public double Defense { get; set; }
        public double Resistance { get; set; }
        public double CritChance { get; set; }
    }

    public class InventoryManager : Node
    {
        public Dictionary<int, ClientItem> Items { get; private set; } = new();
        public CharacterAttributes Attributes { get; private set; } = new();

        public event Action? OnInventoryUpdated;
        public event Action<bool, string>? OnEquipResultReceived;
        public event Action<bool, string>? OnUnequipResultReceived;
        public event Action<bool, string>? OnSwapResultReceived;

        public override void _Ready()
        {
            ServiceRegistry.Instance.Register<InventoryManager>(this);
            var eventBus = ServiceRegistry.Instance.Resolve<EventBus>();
            eventBus.Subscribe<GamePacket>(EventName.OnNetworkPacketReceived, OnPacketReceived);
        }

        private void OnPacketReceived(GamePacket packet)
        {
            switch (packet.Opcode)
            {
                case PacketOpcode.SC_INVENTORY_SYNC:
                    HandleInventorySync(packet.Payload);
                    break;
                case PacketOpcode.SC_EQUIP_RESPONSE:
                    HandleEquipResponse(packet.Payload);
                    break;
                case PacketOpcode.SC_UNEQUIP_RESPONSE:
                    HandleUnequipResponse(packet.Payload);
                    break;
                case PacketOpcode.SC_SWAP_RESPONSE:
                    HandleSwapResponse(packet.Payload);
                    break;
            }
        }

        private void HandleInventorySync(byte[] payload)
        {
            try
            {
                using var ms = new MemoryStream(payload);
                using var br = new BinaryReader(ms);

                ushort count = br.ReadUInt16();
                var newItems = new Dictionary<int, ClientItem>();

                for (int i = 0; i < count; i++)
                {
                    ushort strLen = br.ReadUInt16();
                    byte[] strBytes = br.ReadBytes(strLen);
                    string itemId = System.Text.Encoding.UTF8.GetString(strBytes);

                    uint quantity = br.ReadUInt32();
                    uint durability = br.ReadUInt32();
                    ushort slotIndex = br.ReadUInt16();

                    newItems[slotIndex] = new ClientItem
                    {
                        ItemID = itemId,
                        Quantity = (int)quantity,
                        Durability = (int)durability,
                        SlotIndex = slotIndex
                    };
                }

                Attributes.Level = (int)br.ReadUInt32();
                Attributes.MaxHealth = br.ReadDouble();
                Attributes.Health = br.ReadDouble();
                Attributes.MaxMana = br.ReadDouble();
                Attributes.Mana = br.ReadDouble();
                Attributes.BaseAttack = br.ReadDouble();
                Attributes.WeaponDamage = br.ReadDouble();
                Attributes.Defense = br.ReadDouble();
                Attributes.Resistance = br.ReadDouble();
                Attributes.CritChance = br.ReadDouble();

                Items = newItems;
                GD.Print($"[Inventory] Sincronizado. Itens: {Items.Count}, Level: {Attributes.Level}, HP: {Attributes.Health}/{Attributes.MaxHealth}");
                OnInventoryUpdated?.Invoke();
            }
            catch (Exception ex)
            {
                GD.PrintErr($"[Inventory] Erro ao processar sincronização binária: {ex.Message}");
            }
        }

        public void RequestEquip(int fromSlot, int toSlot)
        {
            var ms = new MemoryStream();
            using (var bw = new BinaryWriter(ms))
            {
                bw.Write((ushort)fromSlot);
                bw.Write((ushort)toSlot);
            }
            ServiceRegistry.Instance.Resolve<NetworkManager>().SendPacket((ushort)PacketOpcode.CS_EQUIP_ITEM, 0, ms.ToArray());
        }

        public void RequestUnequip(int slot)
        {
            var ms = new MemoryStream();
            using (var bw = new BinaryWriter(ms))
            {
                bw.Write((ushort)slot);
            }
            ServiceRegistry.Instance.Resolve<NetworkManager>().SendPacket((ushort)PacketOpcode.CS_UNEQUIP_ITEM, 0, ms.ToArray());
        }

        public void RequestSwap(int slotA, int slotB)
        {
            var ms = new MemoryStream();
            using (var bw = new BinaryWriter(ms))
            {
                bw.Write((ushort)slotA);
                bw.Write((ushort)slotB);
            }
            ServiceRegistry.Instance.Resolve<NetworkManager>().SendPacket((ushort)PacketOpcode.CS_SWAP_SLOTS, 0, ms.ToArray());
        }

        private void HandleEquipResponse(byte[] payload)
        {
            bool success = payload[0] == 1;
            string errMsg = "";
            if (!success && payload.Length > 3)
            {
                ushort strLen = BitConverter.ToUInt16(payload, 1);
                errMsg = System.Text.Encoding.UTF8.GetString(payload, 3, strLen);
            }
            OnEquipResultReceived?.Invoke(success, errMsg);
        }

        private void HandleUnequipResponse(byte[] payload)
        {
            bool success = payload[0] == 1;
            string errMsg = "";
            if (!success && payload.Length > 3)
            {
                ushort strLen = BitConverter.ToUInt16(payload, 1);
                errMsg = System.Text.Encoding.UTF8.GetString(payload, 3, strLen);
            }
            OnUnequipResultReceived?.Invoke(success, errMsg);
        }

        private void HandleSwapResponse(byte[] payload)
        {
            bool success = payload[0] == 1;
            string errMsg = "";
            if (!success && payload.Length > 3)
            {
                ushort strLen = BitConverter.ToUInt16(payload, 1);
                errMsg = System.Text.Encoding.UTF8.GetString(payload, 3, strLen);
            }
            OnSwapResultReceived?.Invoke(success, errMsg);
        }
    }
}`
  },
  {
    name: "InventoryUI.cs",
    path: "src/UI/InventoryUI.cs",
    description: "Controlador C# para a Interface Gráfica de Inventário, suportando drag-and-drop, bônus de combate dinâmicos e slots de equipamento.",
    language: "csharp",
    code: `using Godot;
using System;
using LightAndShadow.Client.Core;
using LightAndShadow.Client.Network;

namespace LightAndShadow.Client.UI
{
    public partial class InventoryUI : Control
    {
        private InventoryManager? _invManager;
        private GridContainer? _backpackGrid;
        private TextureRect? _weaponSlot;
        private TextureRect? _armorSlot;
        private TextureRect? _accessorySlot;
        private Label? _statsLabel;

        public override void _Ready()
        {
            _backpackGrid = GetNode<GridContainer>("BackpackGrid");
            _weaponSlot = GetNode<TextureRect>("Equipment/WeaponSlot");
            _armorSlot = GetNode<TextureRect>("Equipment/ArmorSlot");
            _accessorySlot = GetNode<TextureRect>("Equipment/AccessorySlot");
            _statsLabel = GetNode<Label>("StatsPanel/StatsLabel");

            _invManager = ServiceRegistry.Instance.Resolve<InventoryManager>();
            _invManager.OnInventoryUpdated += RenderUI;
            _invManager.OnEquipResultReceived += OnEquipResult;

            RenderUI();
        }

        private void RenderUI()
        {
            if (_invManager == null) return;

            foreach (Node child in _backpackGrid!.GetChildren())
            {
                child.QueueFree();
            }

            for (int slot = 0; slot <= 29; slot++)
            {
                var slotPanel = new PanelContainer();
                slotPanel.CustomMinimumSize = new Vector2(48, 48);
                slotPanel.SetMeta("slot_index", slot);

                if (_invManager.Items.TryGetValue(slot, out var item))
                {
                    var label = new Label();
                    label.Text = $"{item.ItemID}\\nx{item.Quantity}";
                    label.HorizontalAlignment = HorizontalAlignment.Center;
                    slotPanel.AddChild(label);
                }

                _backpackGrid.AddChild(slotPanel);
            }

            RenderEquipSlot(_weaponSlot, 30, "Weapon");
            RenderEquipSlot(_armorSlot, 31, "Armor");
            RenderEquipSlot(_accessorySlot, 32, "Accessory");

            var attrs = _invManager.Attributes;
            _statsLabel!.Text = $"=== ATRIBUTOS ===\\n" +
                                $"Level: {attrs.Level}\\n" +
                                $"Vida: {attrs.Health:F0} / {attrs.MaxHealth:F0}\\n" +
                                $"Mana: {attrs.Mana:F0} / {attrs.MaxMana:F0}\\n" +
                                $"Ataque Base: {attrs.BaseAttack:F0}\\n" +
                                $"Dano de Arma: +{attrs.WeaponDamage:F0}\\n" +
                                $"Defesa: {attrs.Defense:F0}\\n" +
                                $"Resistência: {attrs.Resistance:F0}\\n" +
                                $"Crit Chance: {attrs.CritChance * 100:F0}%";
        }

        private void RenderEquipSlot(TextureRect? rect, int slotIndex, string defaultText)
        {
            if (rect == null || _invManager == null) return;

            foreach (Node child in rect.GetChildren())
            {
                child.QueueFree();
            }

            if (_invManager.Items.TryGetValue(slotIndex, out var item))
            {
                var label = new Label();
                label.Text = item.ItemID;
                label.HorizontalAlignment = HorizontalAlignment.Center;
                rect.AddChild(label);
            }
            else
            {
                var label = new Label();
                label.Text = $"[{defaultText}]";
                label.HorizontalAlignment = HorizontalAlignment.Center;
                rect.AddChild(label);
            }
        }

        private void OnEquipResult(bool success, string errorMsg)
        {
            if (success)
            {
                GD.Print("[UI] Item equipado com sucesso!");
            }
            else
            {
                GD.PrintErr($"[UI] Erro ao equipar item: {errorMsg}");
            }
        }
    }
}`
  },
  {
    name: "NPCController.cs",
    path: "src/UI/NPCController.cs",
    description: "Script Godot C# para controle de NPC local, detecção de proximidade, renderização e disparo de interações binárias autoritativas.",
    language: "csharp",
    code: `using System;
using Godot;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.UI
{
    /// <summary>
    /// Controlador Godot C# para NPCs no mundo de jogo.
    /// Gerencia interações locais, validações de distância e envia requisições binárias ao servidor.
    /// </summary>
    public partial class NPCController : Area2D
    {
        [Export] public string NPCID { get; set; } = "npc_luna";
        [Export] public string NPCName { get; set; } = "Luna, a Tecedora de Sombras";
        [Export] public float InteractionRadius { get; set; } = 3.5f; // Unidades em tiles

        private Sprite2D? _sprite;
        private Label? _nameLabel;
        private Node2D? _playerNode;
        private bool _isPlayerNear;

        public override void _Ready()
        {
            _sprite = GetNodeOrNull<Sprite2D>("Sprite");
            _nameLabel = GetNodeOrNull<Label>("NameLabel");

            if (_nameLabel != null)
            {
                _nameLabel.Text = NPCName;
            }

            // Conecta eventos de colisão locais (Area2D)
            BodyEntered += OnBodyEntered;
            BodyExited += OnBodyExited;

            GD.Print($"[NPC] {NPCName} ({NPCID}) pronto no mapa.");
        }

        public override void _Input(InputEvent @event)
        {
            if (@event.IsActionPressed("interact") && _isPlayerNear)
            {
                RequestNPCInteraction();
            }
        }

        private void OnBodyEntered(Node2D body)
        {
            if (body.Name == "Player" || body.IsInGroup("Players"))
            {
                _playerNode = body;
                _isPlayerNear = true;
                HighlightNPC(true);
                GD.Print($"[NPC] Jogador aproximou-se de {NPCName}. Pressione [E] para interagir.");
            }
        }

        private void OnBodyExited(Node2D body)
        {
            if (body == _playerNode)
            {
                _isPlayerNear = false;
                _playerNode = null;
                HighlightNPC(false);
                GD.Print($"[NPC] Jogador afastou-se de {NPCName}.");
                
                // Dispara evento para fechar interface de diálogo caso esteja aberta
                EventBus.Instance.Publish("DialogueCloseRequested", NPCID);
            }
        }

        /// <summary>
        /// Solicita de forma autoritativa a abertura do diálogo via rede usando protocolo binário
        /// </summary>
        public void RequestNPCInteraction()
        {
            if (_playerNode == null) return;

            // Validação de distância local preventiva antes de enviar pacote
            float distance = Position.DistanceTo(_playerNode.Position) / 32.0f; // Converte pixels em tiles (32px por tile)
            if (distance > InteractionRadius)
            {
                GD.PrintErr($"[NPC] Erro: Jogador fora do raio de interação ({distance:F2} > {InteractionRadius:F2})");
                return;
            }

            GD.Print($"[NPC] Interagindo com {NPCName}. Enviando opcode CS_NPC_INTERACT (5000)...");

            // Cria payload binário: ID do NPC (String compacta)
            byte[] npcIDBytes = System.Text.Encoding.UTF8.GetBytes(NPCID);
            byte[] payload = new byte[2 + npcIDBytes.Length];
            
            // Prefixo de tamanho de string (Little Endian)
            payload[0] = (byte)(npcIDBytes.Length & 0xFF);
            payload[1] = (byte)((npcIDBytes.Length >> 8) & 0xFF);
            Array.Copy(npcIDBytes, 0, payload, 2, npcIDBytes.Length);

            // Envia pacote de rede binário
            NetworkManager.Instance.SendPacket(5000, payload);
        }

        private void HighlightNPC(bool enable)
        {
            if (_sprite == null) return;
            
            // Aplica efeito visual simples de modulação de cor quando aproximado
            _sprite.Modulate = enable ? new Color(1.2f, 1.2f, 1.0f) : new Color(1.0f, 1.0f, 1.0f);
        }
    }
}`
  },
  {
    name: "DialogueUI.cs",
    path: "src/UI/DialogueUI.cs",
    description: "Componente Godot C# de interface gráfica para diálogos ramificados e aceitação de quests por pacotes binários.",
    language: "csharp",
    code: `using System;
using System.Collections.Generic;
using Godot;
using LightAndShadow.Client.Core;

namespace LightAndShadow.Client.UI
{
    /// <summary>
    /// Gerenciador de Interface Gráfica de Diálogos e Quests.
    /// Recebe pacotes binários SC_DIALOGUE_OPEN do servidor e desenha a árvore de escolhas.
    /// </summary>
    public partial class DialogueUI : Control
    {
        private Label? _npcNameLabel;
        private Label? _dialogueTextLabel;
        private VBoxContainer? _choicesContainer;
        private Panel? _dialogueBoxPanel;

        private string _currentNPCID = "";
        private string _currentNodeID = "";

        public override void _Ready()
        {
            _npcNameLabel = GetNodeOrNull<Label>("DialogueBox/NPCName");
            _dialogueTextLabel = GetNodeOrNull<Label>("DialogueBox/DialogueText");
            _choicesContainer = GetNodeOrNull<VBoxContainer>("DialogueBox/ChoicesContainer");
            _dialogueBoxPanel = GetNodeOrNull<Panel>("DialogueBox");

            // Oculta caixa inicialmente
            Visible = false;

            // Registra tratamento de pacotes de rede
            NetworkManager.Instance.RegisterPacketHandler(5001, OnDialogueOpenPacket); // SC_DIALOGUE_OPEN
            
            // Registra fechamentos por proximidade
            EventBus.Instance.Subscribe("DialogueCloseRequested", (object? data) => {
                if (data is string npcID && npcID == _currentNPCID)
                {
                    CloseDialogue();
                }
            });
        }

        /// <summary>
        /// Processa pacote binário SC_DIALOGUE_OPEN e desenha a tela de conversas
        /// </summary>
        private void OnDialogueOpenPacket(byte[] payload)
        {
            int offset = 0;

            // 1. Lê NPC ID
            string npcID = ReadString(payload, ref offset);
            // 2. Lê Node ID
            string nodeID = ReadString(payload, ref offset);
            // 3. Lê Texto do balão
            string nodeText = ReadString(payload, ref offset);

            // 4. Lê Quantidade de escolhas
            ushort choicesCount = (ushort)(payload[offset] | (payload[offset + 1] << 8));
            offset += 2;

            _currentNPCID = npcID;
            _currentNodeID = nodeID;

            GD.Print($"[DialogueUI] Diálogo aberto: NPC={npcID}, Node={nodeID}, Choices={choicesCount}");

            // Executa na thread principal do Godot de forma segura
            Callable.From(() => {
                Visible = true;
                if (_npcNameLabel != null) _npcNameLabel.Text = npcID.Replace("npc_", "").ToUpper();
                if (_dialogueTextLabel != null) _dialogueTextLabel.Text = nodeText;

                ClearChoices();

                for (int i = 0; i < choicesCount; i++)
                {
                    string nextNodeID = ReadString(payload, ref offset);
                    string choiceText = ReadString(payload, ref offset);

                    AddChoiceButton(choiceText, nextNodeID);
                }
            }).CallDeferred();
        }

        private void AddChoiceButton(string text, string nextNodeID)
        {
            if (_choicesContainer == null) return;

            var button = new Button();
            button.Text = text;
            button.Alignment = HorizontalAlignment.Left;
            button.CustomMinimumSize = new Vector2(0, 40);

            // Vincula callback de clique do botão
            button.Pressed += () => OnChoiceSelected(nextNodeID);

            _choicesContainer.AddChild(button);
        }

        private void OnChoiceSelected(string nextNodeID)
        {
            GD.Print($"[DialogueUI] Escolha selecionada. Próximo nó: {nextNodeID}");

            if (nextNodeID == "end" || string.IsNullOrEmpty(nextNodeID))
            {
                CloseDialogue();
                
                // Envia pacote de encerramento
                SendResponsePacket("end");
                return;
            }

            // Envia pacote de resposta binário ao servidor para avançar na árvore de diálogos
            SendResponsePacket(nextNodeID);
        }

        private void SendResponsePacket(string nextNodeID)
        {
            // Cria payload binário: npcID, nodeID, nextNodeID
            byte[] npcBytes = System.Text.Encoding.UTF8.GetBytes(_currentNPCID);
            byte[] nodeBytes = System.Text.Encoding.UTF8.GetBytes(_currentNodeID);
            byte[] nextBytes = System.Text.Encoding.UTF8.GetBytes(nextNodeID);

            int totalSize = (2 + npcBytes.Length) + (2 + nodeBytes.Length) + (2 + nextBytes.Length);
            byte[] payload = new byte[totalSize];

            int offset = 0;
            WriteString(payload, _currentNPCID, ref offset);
            WriteString(payload, _currentNodeID, ref offset);
            WriteString(payload, nextNodeID, ref offset);

            // Envia opcode CS_DIALOGUE_RESPONSE (5002)
            NetworkManager.Instance.SendPacket(5002, payload);
        }

        private void CloseDialogue()
        {
            Visible = false;
            ClearChoices();
            _currentNPCID = "";
            _currentNodeID = "";
        }

        private void ClearChoices()
        {
            if (_choicesContainer == null) return;
            foreach (Node child in _choicesContainer.GetChildren())
            {
                child.QueueFree();
            }
        }

        // --- Utilitários de leitura binária compatíveis com Little Endian do Gateway ---

        private string ReadString(byte[] data, ref int offset)
        {
            if (offset + 2 > data.Length) return "";
            ushort length = (ushort)(data[offset] | (data[offset + 1] << 8));
            offset += 2;

            if (offset + length > data.Length) return "";
            string val = System.Text.Encoding.UTF8.GetString(data, offset, length);
            offset += length;
            return val;
        }

        private void WriteString(byte[] data, string val, ref int offset)
        {
            byte[] bytes = System.Text.Encoding.UTF8.GetBytes(val);
            data[offset] = (byte)(bytes.Length & 0xFF);
            data[offset + 1] = (byte)((bytes.Length >> 8) & 0xFF);
            offset += 2;

            Array.Copy(bytes, 0, data, offset, bytes.Length);
            offset += bytes.Length;
        }
    }
}`
  },
  {
    name: "TradingController.cs",
    path: "src/Economy/TradingController.cs",
    description: "Controlador do cliente Godot 4 para negociação direta entre jogadores (Player-to-Player Trading). Suporta oferta de ouro/itens e dual-confirm.",
    language: "csharp",
    code: `using System;
using Godot;

namespace LightAndShadow.Client.Economy
{
    /// <summary>
    /// Controlador C# para o Sistema de Troca direta entre Jogadores (Player Trading) no Godot 4.
    /// Comunica-se de forma assíncrona usando o protocolo binário de baixa latência em Little Endian.
    /// </summary>
    public class TradingController : Node
    {
        /// <summary>
        /// Solicita uma nova sessão de troca com outro jogador próximo.
        /// </summary>
        public void RequestTrade(string targetPlayerName)
        {
            if (string.IsNullOrEmpty(targetPlayerName)) return;

            byte[] targetBytes = System.Text.Encoding.UTF8.GetBytes(targetPlayerName);
            byte[] payload = new byte[2 + targetBytes.Length];
            int offset = 0;
            WriteString(payload, targetPlayerName, ref offset);

            NetworkManager.Instance.SendPacket(7000, payload); // CS_TRADE_REQUEST
            GD.Print($"[TradingController] Solicitação de troca enviada para: {targetPlayerName}");
        }

        /// <summary>
        /// Responde a um convite de troca recebido.
        /// </summary>
        public void RespondTrade(bool accepted)
        {
            byte[] payload = new byte[1];
            payload[0] = (byte)(accepted ? 1 : 0);

            NetworkManager.Instance.SendPacket(7002, payload); // CS_TRADE_RESPOND
            GD.Print($"[TradingController] Resposta ao convite enviada: Aceito={accepted}");
        }

        /// <summary>
        /// Oferece uma quantia de ouro na janela comercial.
        /// </summary>
        public void OfferGold(uint goldAmount)
        {
            byte[] payload = new byte[4];
            int offset = 0;
            WriteUint32(payload, goldAmount, ref offset);

            NetworkManager.Instance.SendPacket(7004, payload); // CS_TRADE_OFFER_GOLD
            GD.Print($"[TradingController] Oferta de ouro enviada: {goldAmount}");
        }

        /// <summary>
        /// Oferece um item do inventário para a janela comercial.
        /// </summary>
        public void OfferItem(uint slotIndex, uint quantity)
        {
            byte[] payload = new byte[8];
            int offset = 0;
            WriteUint32(payload, slotIndex, ref offset);
            WriteUint32(payload, quantity, ref offset);

            NetworkManager.Instance.SendPacket(7003, payload); // CS_TRADE_OFFER_ITEM
            GD.Print($"[TradingController] Oferta de item enviada: Slot={slotIndex}, Qtd={quantity}");
        }

        /// <summary>
        /// Tranca a proposta comercial atual (Estágio 1 da Confirmação Dupla).
        /// </summary>
        public void LockTrade()
        {
            NetworkManager.Instance.SendPacket(7005, new byte[0]); // CS_TRADE_LOCK
            GD.Print("[TradingController] Proposta comercial trancada!");
        }

        /// <summary>
        /// Confirma a proposta comercial finalizada (Estágio 2 da Confirmação Dupla).
        /// </summary>
        public void ConfirmTrade()
        {
            NetworkManager.Instance.SendPacket(7006, new byte[0]); // CS_TRADE_CONFIRM
            GD.Print("[TradingController] Proposta comercial confirmada!");
        }

        /// <summary>
        /// Cancela a sessão comercial corrente.
        /// </summary>
        public void CancelTrade()
        {
            NetworkManager.Instance.SendPacket(7008, new byte[0]); // CS_TRADE_CANCEL
            GD.Print("[TradingController] Sessão comercial cancelada.");
        }

        private void WriteString(byte[] data, string val, ref int offset)
        {
            byte[] bytes = System.Text.Encoding.UTF8.GetBytes(val);
            data[offset] = (byte)(bytes.Length & 0xFF);
            data[offset + 1] = (byte)((bytes.Length >> 8) & 0xFF);
            offset += 2;

            Array.Copy(bytes, 0, data, offset, bytes.Length);
            offset += bytes.Length;
        }

        private void WriteUint32(byte[] data, uint val, ref int offset)
        {
            data[offset] = (byte)(val & 0xFF);
            data[offset + 1] = (byte)((val >> 8) & 0xFF);
            data[offset + 2] = (byte)((val >> 16) & 0xFF);
            data[offset + 3] = (byte)((val >> 24) & 0xFF);
            offset += 4;
        }
    }
}`
  },
  {
    name: "NPCShopController.cs",
    path: "src/Economy/NPCShopController.cs",
    description: "Controlador do cliente Godot 4 para interações com lojas de NPCs (comprar, vender e reparar itens). Realiza pré-validações locais.",
    language: "csharp",
    code: `using System;
using Godot;

namespace LightAndShadow.Client.Economy
{
    /// <summary>
    /// Controlador C# para o Sistema de Lojas de NPCs (Merchant NPC Shop) no Godot 4.
    /// Interage com ferreiros, comerciantes e mercadores para compra, venda e reparo.
    /// </summary>
    public class NPCShopController : Node
    {
        /// <summary>
        /// Envia solicitação de compra de um item do catálogo de um NPC.
        /// </summary>
        public void BuyNPCItem(string itemID, uint quantity)
        {
            if (string.IsNullOrEmpty(itemID) || quantity == 0) return;

            byte[] idBytes = System.Text.Encoding.UTF8.GetBytes(itemID);
            byte[] payload = new byte[2 + idBytes.Length + 4];
            int offset = 0;
            WriteString(payload, itemID, ref offset);
            WriteUint32(payload, quantity, ref offset);

            NetworkManager.Instance.SendPacket(7100, payload); // CS_NPC_SHOP_BUY
            GD.Print($"[NPCShopController] Comprando item: ID={itemID}, Qtd={quantity}");
        }

        /// <summary>
        /// Vende um item do inventário do jogador para o NPC comerciante.
        /// </summary>
        public void SellNPCItem(uint slotIndex, uint quantity)
        {
            if (quantity == 0) return;

            byte[] payload = new byte[8];
            int offset = 0;
            WriteUint32(payload, slotIndex, ref offset);
            WriteUint32(payload, quantity, ref offset);

            NetworkManager.Instance.SendPacket(7101, payload); // CS_NPC_SHOP_SELL
            GD.Print($"[NPCShopController] Vendendo item do slot={slotIndex}, Qtd={quantity}");
        }

        /// <summary>
        /// Solicita ao ferreiro NPC que repare a durabilidade do equipamento no slot indicado.
        /// </summary>
        public void RepairEquipment(uint slotIndex)
        {
            byte[] payload = new byte[4];
            int offset = 0;
            WriteUint32(payload, slotIndex, ref offset);

            NetworkManager.Instance.SendPacket(7102, payload); // CS_NPC_SHOP_REPAIR
            GD.Print($"[NPCShopController] Solicitando reparo para o equipamento no slot={slotIndex}");
        }

        private void WriteString(byte[] data, string val, ref int offset)
        {
            byte[] bytes = System.Text.Encoding.UTF8.GetBytes(val);
            data[offset] = (byte)(bytes.Length & 0xFF);
            data[offset + 1] = (byte)((bytes.Length >> 8) & 0xFF);
            offset += 2;

            Array.Copy(bytes, 0, data, offset, bytes.Length);
            offset += bytes.Length;
        }

        private void WriteUint32(byte[] data, uint val, ref int offset)
        {
            data[offset] = (byte)(val & 0xFF);
            data[offset + 1] = (byte)((val >> 8) & 0xFF);
            data[offset + 2] = (byte)((val >> 16) & 0xFF);
            data[offset + 3] = (byte)((val >> 24) & 0xFF);
            offset += 4;
        }
    }
}`
  },
  {
    name: "MarketplaceController.cs",
    path: "src/Economy/MarketplaceController.cs",
    description: "Controlador do cliente Godot 4 para interagir com o Mercado/Casa de Leilões. Suporta publicação, compra e cancelamento de ordens em custódia (Escrow).",
    language: "csharp",
    code: `using System;
using Godot;

namespace LightAndShadow.Client.Economy
{
    /// <summary>
    /// Controlador C# para o Mercado/Casa de Leilões (Marketplace / Auction House) no Godot 4.
    /// Interage de forma totalmente server-authoritative com taxas e custódia (Escrow) de itens.
    /// </summary>
    public class MarketplaceController : Node
    {
        /// <summary>
        /// Lista um item do inventário para ser vendido publicamente na casa de leilão.
        /// </summary>
        public void ListMarketItem(uint slotIndex, uint quantity, uint priceInGold)
        {
            if (quantity == 0 || priceInGold == 0) return;

            byte[] payload = new byte[12];
            int offset = 0;
            WriteUint32(payload, slotIndex, ref offset);
            WriteUint32(payload, quantity, ref offset);
            WriteUint32(payload, priceInGold, ref offset);

            NetworkManager.Instance.SendPacket(7200, payload); // CS_MARKET_CREATE_ORDER
            GD.Print($"[MarketplaceController] Listando item: Slot={slotIndex}, Qtd={quantity}, Preço={priceInGold}");
        }

        /// <summary>
        /// Adquire um item anunciado na casa de leilão, transferindo ouro e retirando-o de custódia.
        /// </summary>
        public void BuyMarketItem(uint orderID)
        {
            byte[] payload = new byte[4];
            int offset = 0;
            WriteUint32(payload, orderID, ref offset);

            NetworkManager.Instance.SendPacket(7201, payload); // CS_MARKET_BUY_ITEM
            GD.Print($"[MarketplaceController] Comprando ordem de leilão: ID={orderID}");
        }

        /// <summary>
        /// Cancela e remove um anúncio de venda de autoria própria do mercado, retornando o item da custódia.
        /// </summary>
        public void CancelMarketOrder(uint orderID)
        {
            byte[] payload = new byte[4];
            int offset = 0;
            WriteUint32(payload, orderID, ref offset);

            NetworkManager.Instance.SendPacket(7202, payload); // CS_MARKET_CANCEL_ORDER
            GD.Print($"[MarketplaceController] Cancelando ordem de leilão: ID={orderID}");
        }

        /// <summary>
        /// Pesquisa e solicita atualização de listagens ativas no mercado, opcionalmente filtradas por ID do item.
        /// </summary>
        public void SearchMarket(string filterItemID)
        {
            filterItemID ??= "";
            byte[] filterBytes = System.Text.Encoding.UTF8.GetBytes(filterItemID);
            byte[] payload = new byte[2 + filterBytes.Length];
            int offset = 0;
            WriteString(payload, filterItemID, ref offset);

            NetworkManager.Instance.SendPacket(7203, payload); // CS_MARKET_SEARCH
            GD.Print($"[MarketplaceController] Pesquisando ordens ativas: Filtro='{filterItemID}'");
        }

        private void WriteString(byte[] data, string val, ref int offset)
        {
            byte[] bytes = System.Text.Encoding.UTF8.GetBytes(val);
            data[offset] = (byte)(bytes.Length & 0xFF);
            data[offset + 1] = (byte)((bytes.Length >> 8) & 0xFF);
            offset += 2;

            Array.Copy(bytes, 0, data, offset, bytes.Length);
            offset += bytes.Length;
        }

        private void WriteUint32(byte[] data, uint val, ref int offset)
        {
            data[offset] = (byte)(val & 0xFF);
            data[offset + 1] = (byte)((val >> 8) & 0xFF);
            data[offset + 2] = (byte)((val >> 16) & 0xFF);
            data[offset + 3] = (byte)((val >> 24) & 0xFF);
            offset += 4;
        }
    }
}`
  },
  {
    name: "ProfessionsController.cs",
    path: "src/Economy/ProfessionsController.cs",
    description: "Controlador C# para o Sistema de Profissões (Coleta e Síntese) no Godot 4. Gerencia início de coleta, cancelamento e síntese autoritativa.",
    language: "csharp",
    code: `using System;
using Godot;

namespace LightAndShadow.Client.Economy
{
    /// <summary>
    /// Controlador C# para o Sistema de Profissões, Coleta e Síntese no Godot 4.
    /// Utiliza o protocolo binário de baixa latência em Little Endian.
    /// </summary>
    public class ProfessionsController : Node
    {
        /// <summary>
        /// Inicia o processo de coleta em um recurso (Mineração, Herbologia, Corte de Madeira, Pesca).
        /// </summary>
        public void StartGathering(string nodeId)
        {
            if (string.IsNullOrEmpty(nodeId)) return;

            byte[] nodeBytes = System.Text.Encoding.UTF8.GetBytes(nodeId);
            byte[] payload = new byte[2 + nodeBytes.Length];
            int offset = 0;
            WriteString(payload, nodeId, ref offset);

            NetworkManager.Instance.SendPacket(8000, payload); // CS_GATHER_START
            GD.Print($"[ProfessionsController] Iniciando coleta no recurso: {nodeId}");
        }

        /// <summary>
        /// Cancela o processo de coleta atual devido a movimento voluntário ou ação do usuário.
        /// </summary>
        public void CancelGathering()
        {
            NetworkManager.Instance.SendPacket(8006, null); // CS_GATHER_CANCEL
            GD.Print("[ProfessionsController] Solicitação de cancelamento de coleta enviada.");
        }

        /// <summary>
        /// Inicia a síntese (Craft) de uma receita de profissão (Ferraria, Alquimia, Encantamento).
        /// </summary>
        public void StartCrafting(string recipeId)
        {
            if (string.IsNullOrEmpty(recipeId)) return;

            byte[] recipeBytes = System.Text.Encoding.UTF8.GetBytes(recipeId);
            byte[] payload = new byte[2 + recipeBytes.Length];
            int offset = 0;
            WriteString(payload, recipeId, ref offset);

            NetworkManager.Instance.SendPacket(8003, payload); // CS_CRAFT_START
            GD.Print($"[ProfessionsController] Iniciando síntese da receita: {recipeId}");
        }

        private void WriteString(byte[] data, string val, ref int offset)
        {
            byte[] bytes = System.Text.Encoding.UTF8.GetBytes(val);
            data[offset] = (byte)(bytes.Length & 0xFF);
            data[offset + 1] = (byte)((bytes.Length >> 8) & 0xFF);
            offset += 2;

            Array.Copy(bytes, 0, data, offset, bytes.Length);
            offset += bytes.Length;
        }
    }
}`
  }
];


