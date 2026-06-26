using System;
using System.Threading;
using System.Windows.Forms;

namespace controller_mcp.TestTarget
{
    static class Program
    {
        // Deterministic Memory Variables for MemoryTools testing
        public static int[] DynamicMemory = new int[1024];
        public static string TestString = "YADMS_MEMORY_TEST_STRING";
        public static bool KeepRunning = true;

        [STAThread]
        static void Main(string[] args)
        {
            // Allocate the target into a dynamic heap array so it lands in PAGE_READWRITE
            DynamicMemory[512] = 1337;
            if (args.Length > 0 && args[0] == "--hidden")
            {
                // Run without showing UI
                while (KeepRunning)
                {
                    Thread.Sleep(100);
                }
                return;
            }

            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            
            Form form = new Form
            {
                Text = "YADMS_TEST_WINDOW",
                Width = 400,
                Height = 300
            };
            
            // Expose handle natively for testing
            form.HandleCreated += (s, e) => {
                Console.WriteLine($"HWND:{form.Handle}");
            };

            Application.Run(form);
        }
    }
}
