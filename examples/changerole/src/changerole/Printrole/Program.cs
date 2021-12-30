using Microsoft.ServiceFabric.Services.Runtime;
using System;
using System.Fabric;
using System.Threading;
using System.Threading.Tasks;

namespace changerole
{
    class Program
    {
        private class Dummy : StatefulService
        {
            public Dummy(StatefulServiceContext serviceContext) : base(serviceContext)
            {
            }

            protected override void OnAbort()
            {
                Console.WriteLine("on abort");
                base.OnAbort();
            }

            protected override Task OnChangeRoleAsync(ReplicaRole newRole, CancellationToken cancellationToken)
            {
                Console.WriteLine("change to " + newRole);
                return base.OnChangeRoleAsync(newRole, cancellationToken);
            }

            protected override Task OnCloseAsync(CancellationToken cancellationToken)
            {
                Console.WriteLine("on close");
                return base.OnCloseAsync(cancellationToken);
            }

            protected override Task OnOpenAsync(ReplicaOpenMode openMode, CancellationToken cancellationToken)
            {
                Console.WriteLine("on open");
                return base.OnOpenAsync(openMode, cancellationToken);
            }

            protected override async Task RunAsync(CancellationToken cancellationToken)
            {
                Console.WriteLine("on run");
            }
        }

        static void Main(string[] args)
        {
            Console.WriteLine("starting print change role");
            ServiceRuntime.RegisterServiceAsync("PrintroleType", ctx => new Dummy(ctx)).GetAwaiter().GetResult();
            Console.WriteLine("registered");
            Thread.Sleep(Timeout.Infinite);
        }
    }
}
