import { useScanner } from "./useScanner";
import { ScannerView } from "./ScannerPage.view";

export function ScannerPage() {
  const state = useScanner();
  return <ScannerView {...state} />;
}

export default ScannerPage;
