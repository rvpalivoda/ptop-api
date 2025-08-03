import { useEffect, useState } from "react";
import { X } from "lucide-react";
import { createAdvert } from "@/api/adverts";
import { useAuth } from "@/context/AuthContext";
import {
  getAssets,
  getCountries,
  getDurations,
  getPaymentMethods,
} from "@/api/dictionaries";
import { toast } from "@/components/ui/sonner";

interface CreateAdvertFormProps {
  onClose: () => void;
  type: "buy" | "sell";
}

export const CreateAdvertForm = ({ onClose, type }: CreateAdvertFormProps) => {
  const { userInfo } = useAuth();
  const [quoteAssets, setQuoteAssets] = useState<string[]>([]);
  const [formData, setFormData] = useState({
    baseAsset: "",
    quoteAsset: "",
    amount: "",
    price: "",
    country: "",
    postCode: "",
    paymentMethods: [] as { name: string; method_name: string }[],
    visibility: "public",
    terms: "",
    minAmount: "",
    maxAmount: "",
    duration: "",
    status: "Online",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const [assets, setAssets] = useState<string[]>([]);
  const [countries, setCountries] = useState<string[]>([]);
  const [durations, setDurations] = useState<
    {
      id: string;
      label: string;
      isDefault?: boolean;
    }[]
  >([]);
  const [paymentMethods, setPaymentMethods] = useState<
    { name: string; method_name: string }[]
  >([]);

  useEffect(() => {
    async function loadDictionaries() {
      try {
        const a = await getAssets();
        const mappedAssets = a.map((x: any) => x.asset_code ?? x.name ?? x);
        setAssets(mappedAssets);
        setQuoteAssets(mappedAssets);
        const c = await getCountries();
        setCountries(c.map((x: any) => x.name ?? x));
        const d = await getDurations();
        setDurations(
          d.map((x: any) => ({
            id: x.name ?? x,
            label: x.duration_name ?? x,
            isDefault: x.is_default ?? false,
          })),
        );
        if (a.length) {
          setFormData((f) => ({
            ...f,
            baseAsset: f.baseAsset || a[0].asset_code || a[0],
            quoteAsset:
              f.quoteAsset ||
              mappedAssets.find(
                (x: string) => x !== (f.baseAsset || a[0].asset_code || a[0]),
              ) ||
              mappedAssets[0],
          }));
        }
      } catch (err) {
        console.error("load dictionaries error:", err);
      }
    }
    loadDictionaries();
  }, []);

  useEffect(() => {
    if (durations.length && !formData.duration) {
      const def = durations.find((d) => d.isDefault) ?? durations[0];
      setFormData((f) => ({ ...f, duration: def.id }));
    }
  }, [durations]);

  useEffect(() => {
    if (formData.baseAsset && formData.baseAsset === formData.quoteAsset) {
      const qa = quoteAssets.find((a) => a !== formData.baseAsset) || "";
      setFormData((f) => ({ ...f, quoteAsset: qa }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData.baseAsset, quoteAssets]);

  useEffect(() => {
    if (!formData.country) {
      setPaymentMethods([]);
      return;
    }
    async function loadPaymentMethods() {
      try {
        const methods = await getPaymentMethods(formData.country);
        setPaymentMethods(
          methods.map((x: any) => ({
            name: x.name ?? x,
            method_name: x.method_name ?? x.name ?? x,
          })),
        );
        setFormData((f) => ({ ...f, paymentMethods: [] }));
      } catch (err) {
        console.error("load payment methods error:", err);
      }
    }
    loadPaymentMethods();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData.country]);

  const validate = () => {
    const e: Record<string, string> = {};
    if (!formData.baseAsset) e.baseAsset = "Обязательное поле";
    if (!formData.quoteAsset) e.quoteAsset = "Обязательное поле";
    if (formData.baseAsset && formData.baseAsset === formData.quoteAsset)
      e.quoteAsset = "Базовый и котируемый актив должны различаться";
    if (!formData.amount) e.amount = "Обязательное поле";
    if (!formData.price) e.price = "Обязательное поле";
    if (!formData.country) e.country = "Обязательное поле";
    if (!formData.postCode) e.postCode = "Обязательное поле";
    if (!formData.minAmount) e.minAmount = "Обязательное поле";
    if (!formData.maxAmount) e.maxAmount = "Обязательное поле";
    if (!formData.duration) e.duration = "Обязательное поле";
    if (formData.paymentMethods.length === 0)
      e.paymentMethods = "Выберите хотя бы один способ оплаты";
    return e;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const eMap = validate();
    if (Object.keys(eMap).length) {
      setErrors(eMap);
      return;
    }
    try {
      await createAdvert({
        client: userInfo?.username || "",
        status: formData.status,
        type: type === "buy" ? "Buy" : "Sell",
        base_asset: formData.baseAsset,
        quote_asset: formData.quoteAsset,
        amount: Number(formData.amount),
        price: Number(formData.price),
        country: formData.country,
        post_code: formData.postCode,
        payment_method: formData.paymentMethods.map((pm) => ({
          payment_method: pm.name,
          payment_method_name: pm.method_name,
        })),
        visibility_status: formData.visibility as
          | "public"
          | "private"
          | "private-by-link",
        terms: formData.terms,
        min_amount: Number(formData.minAmount),
        max_amount: Number(formData.maxAmount),
        duration: formData.duration,
      });
      toast("Объявление создано");
      onClose();
    } catch (err) {
      console.error("Create advert error:", err);
      toast(err instanceof Error ? err.message : "Ошибка создания");
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-gray-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-white">
            {type === "buy"
              ? "Создать объявление о покупке"
              : "Создать объявление о продаже"}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Криптовалюта
              </label>
              <select
                value={formData.baseAsset}
                onChange={(e) => {
                  const v = e.target.value;
                  setFormData((f) => {
                    let qa = f.quoteAsset;
                    if (qa === v) {
                      qa = quoteAssets.find((a) => a !== v) || "";
                    }
                    return { ...f, baseAsset: v, quoteAsset: qa };
                  });
                }}
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {assets.map((a) => (
                  <option key={a} value={a}>
                    {a}
                  </option>
                ))}
              </select>
              {errors.baseAsset && (
                <p className="text-sm text-red-500 mt-1">{errors.baseAsset}</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Валюта расчёта
              </label>
              <select
                value={formData.quoteAsset}
                onChange={(e) =>
                  setFormData({ ...formData, quoteAsset: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {quoteAssets
                  .filter((a) => a !== formData.baseAsset)
                  .map((a) => (
                    <option key={a} value={a}>
                      {a}
                    </option>
                  ))}
              </select>
              {errors.quoteAsset && (
                <p className="text-sm text-red-500 mt-1">{errors.quoteAsset}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Цена
              </label>
              <input
                type="number"
                value={formData.price}
                onChange={(e) =>
                  setFormData({ ...formData, price: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              {errors.price && (
                <p className="text-sm text-red-500 mt-1">{errors.price}</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Страна
              </label>
              <select
                value={formData.country}
                onChange={(e) =>
                  setFormData({ ...formData, country: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {countries.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
          {errors.country && (
            <p className="text-sm text-red-500 mt-1">{errors.country}</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Объём
          </label>
          <input
            type="number"
            value={formData.amount}
            onChange={(e) =>
              setFormData({ ...formData, amount: e.target.value })
            }
            className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {errors.amount && (
            <p className="text-sm text-red-500 mt-1">{errors.amount}</p>
          )}
        </div>
      </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Почтовый индекс
              </label>
              <input
                type="text"
                value={formData.postCode}
                onChange={(e) =>
                  setFormData({ ...formData, postCode: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              {errors.postCode && (
                <p className="text-sm text-red-500 mt-1">{errors.postCode}</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Длительность
              </label>
              <select
                value={formData.duration}
                onChange={(e) =>
                  setFormData({ ...formData, duration: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {durations.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.label}
                  </option>
                ))}
              </select>
              {errors.duration && (
                <p className="text-sm text-red-500 mt-1">{errors.duration}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Мин. сумма
              </label>
              <input
                type="number"
                value={formData.minAmount}
                onChange={(e) =>
                  setFormData({ ...formData, minAmount: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              {errors.minAmount && (
                <p className="text-sm text-red-500 mt-1">{errors.minAmount}</p>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Макс. сумма
              </label>
              <input
                type="number"
                value={formData.maxAmount}
                onChange={(e) =>
                  setFormData({ ...formData, maxAmount: e.target.value })
                }
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              {errors.maxAmount && (
                <p className="text-sm text-red-500 mt-1">{errors.maxAmount}</p>
              )}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Способы оплаты
            </label>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {paymentMethods.map((method) => (
                <label
                  key={method.name}
                  className="flex items-center space-x-2 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    className="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
                    onChange={(e) => {
                      if (e.target.checked) {
                        setFormData({
                          ...formData,
                          paymentMethods: [...formData.paymentMethods, method],
                        });
                      } else {
                        setFormData({
                          ...formData,
                          paymentMethods: formData.paymentMethods.filter(
                            (m) => m.name !== method.name,
                          ),
                        });
                      }
                    }}
                  />
                  <span className="text-gray-300 text-sm">
                    {method.method_name}
                  </span>
                </label>
              ))}
            </div>
            {errors.paymentMethods && (
              <p className="text-sm text-red-500 mt-1">
                {errors.paymentMethods}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Условия
            </label>
            <textarea
              value={formData.terms}
              onChange={(e) =>
                setFormData({ ...formData, terms: e.target.value })
              }
              rows={4}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="flex space-x-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-3 px-4 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
            >
              Отмена
            </button>
            <button
              type="submit"
              className={`flex-1 py-3 px-4 rounded-lg font-medium transition-colors ${
                type === "buy"
                  ? "bg-green-600 hover:bg-green-700 text-white"
                  : "bg-red-600 hover:bg-red-700 text-white"
              }`}
            >
              Создать объявление
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
